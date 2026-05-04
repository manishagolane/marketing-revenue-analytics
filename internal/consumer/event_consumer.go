package consumer

import (
	"context"
	"encoding/json"
	"marketing-revenue-analytics/internal/events"
	"marketing-revenue-analytics/models"
	"marketing-revenue-analytics/utils"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type EventConsumer struct {
	queries *models.Queries
	logger  *zap.Logger
	conn    *pgxpool.Pool
}

func NewEventConsumer(q *models.Queries, conn *pgxpool.Pool, logger *zap.Logger) *EventConsumer {
	return &EventConsumer{
		queries: q,
		logger:  logger,
		conn:    conn,
	}
}

func (c *EventConsumer) Consume(message string) error {
	var event events.EventPayload

	err := json.Unmarshal([]byte(message), &event)
	if err != nil {
		c.logger.Error("invalid message", zap.Error(err))
		return err
	}

	ctx := context.Background()

	metadata, err := utils.ToJSONB(event.Metadata)
	if err != nil {
		c.logger.Error("failed to serialize metadata", zap.Error(err))
		return err
	}

	tx, err := c.conn.Begin(ctx)
	if err != nil {
		c.logger.Error("failed to begin tx", zap.Error(err))
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	qtx := c.queries.WithTx(tx)

	// If this returns a duplicate key error → event already processed
	// Skip aggregation, return nil → SQS deletes the message cleanly
	_, err = qtx.CreateEventLog(ctx, models.CreateEventLogParams{
		ID:         event.ID,
		CampaignID: event.CampaignID,
		EventType:  event.EventType,
		SourceUrl:  utils.ToNullString(event.SourceURL),
		IpAddress:  utils.ToNullString(event.IP),
		UserAgent:  utils.ToNullString(event.UserAgent),
		Metadata:   metadata,
		OccurredAt: event.OccurredAt,
		SessionID:  utils.ToNullString(event.SessionID),
		Step:       utils.ToNullString(event.Step),
	})
	if err != nil {
		if isDuplicateKeyError(err) {
			c.logger.Warn("duplicate event skipped",
				zap.String("event_id", event.ID),
			)
			return nil // delete from SQS
		}

		if isValidationError(err) {
			c.logger.Warn("invalid event → sending to DLQ",
				zap.String("event_id", event.ID),
				zap.String("event_type", event.EventType),
				zap.Error(err),
			)
			return err //  allow retry → then DLQ
		}

		c.logger.Error("failed to insert event log", zap.Error(err))
		return err
	}

	//  increment daily aggregate
	impressions, clicks, conversions := eventCounts(event.EventType)

	// Use time.Time directly — SQLC maps DATE to time.Time
	metricDate := event.OccurredAt.UTC().Truncate(24 * time.Hour)

	if err := qtx.UpsertDailyMetrics(ctx, models.UpsertDailyMetricsParams{
		CampaignID:  event.CampaignID,
		Date:        metricDate, // time.Time, not string
		Impressions: impressions,
		Clicks:      clicks,
		Conversions: conversions,
	}); err != nil {
		c.logger.Error("failed to upsert daily metrics", zap.Error(err))
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		c.logger.Error("commit failed", zap.Error(err))
		return err
	}

	c.logger.Info("event consumed",
		zap.String("event_id", event.ID),
		zap.String("campaign_id", event.CampaignID),
		zap.String("type", event.EventType),
	)
	return nil
}

func eventCounts(eventType string) (int32, int32, int32) {
	switch eventType {
	case "impression":
		return 1, 0, 0
	case "click":
		return 0, 1, 0
	case "conversion":
		return 0, 0, 1
	default:
		return 0, 0, 0
	}
}

// isDuplicateKeyError checks for Postgres unique violation (code 23505)
func isDuplicateKeyError(err error) bool {
	return strings.Contains(err.Error(), "23505") ||
		strings.Contains(err.Error(), "duplicate key")
}

func isValidationError(err error) bool {
	return strings.Contains(err.Error(), "event_logs_event_type_check") ||
		strings.Contains(err.Error(), "violates check constraint")
}
