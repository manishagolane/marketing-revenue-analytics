package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"

	"marketing-revenue-analytics/config"
	"marketing-revenue-analytics/internal/dto"
	"marketing-revenue-analytics/internal/events"
	"marketing-revenue-analytics/utils"
)

type EventHandler struct {
	queries EventStore
	clients QueueStore
	logger  *zap.Logger
}

func NewEventHandler(q EventStore, clients QueueStore, logger *zap.Logger) *EventHandler {
	return &EventHandler{
		queries: q,
		clients: clients,
		logger:  logger,
	}
}

type TrackEventRequest struct {
	CampaignID string                 `json:"campaign_id" binding:"required"`
	EventType  string                 `json:"event_type"  binding:"required,oneof=impression click conversion"`
	SourceURL  string                 `json:"source_url"`
	OccurredAt *string                `json:"occurred_at"` // RFC3339, optional — defaults to NOW
	Metadata   map[string]interface{} `json:"metadata"`
	SessionID  string                 `json:"session_id"` // optional — groups events into a session
	Step       string                 `json:"step"        binding:"omitempty,oneof=ad landing signup purchase"`
}

// POST /api/v1/events/track
// Public endpoint — external systems send events here, no JWT required
func (h *EventHandler) Track(c *gin.Context) {
	var req TrackEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendBadRequestError(c, err)
		return
	}

	ctx := c.Request.Context()

	// Verify campaign exists and is active
	campaign, err := h.queries.GetCampaignByID(ctx, req.CampaignID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			SendNotFoundError(c, errors.New("campaign not found"))
			return
		}
		SendApplicationError(c, err)
		return
	}

	if campaign.Status != "active" {
		SendBadRequestError(c, errors.New("campaign is not active"))
		return
	}

	// Parse occurred_at — default to now if not provided
	occurredAt := time.Now().UTC()
	if req.OccurredAt != nil && *req.OccurredAt != "" {
		t, err := time.Parse(time.RFC3339, *req.OccurredAt)
		if err != nil {
			SendBadRequestError(c, errors.New("invalid occurred_at, use RFC3339"))
			return
		}
		occurredAt = t
	}

	events_id, ulidErr := utils.GetUlid()
	if ulidErr != nil {
		h.logger.Error("failed to generate ULID", zap.Error(ulidErr))
		SendApplicationError(c, ulidErr)
		return
	}
	payload := events.EventPayload{
		ID:         events_id,
		CampaignID: req.CampaignID,
		EventType:  req.EventType,
		SourceURL:  req.SourceURL,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		Metadata:   req.Metadata,
		OccurredAt: occurredAt,
		SessionID:  req.SessionID,
		Step:       req.Step,
	}

	queueUrl := config.GetString("aws.sqs")
	h.logger.Debug("queueUrl", zap.String("queueUrl:", queueUrl))

	err = h.clients.SendSqsMessageWithBody(ctx, payload, queueUrl)
	if err != nil {
		SendApplicationError(c, err)
		return
	}

	// 202 Accepted — event received, will be processed
	c.JSON(http.StatusAccepted, dto.APIResponse{
		Status:  "success",
		Message: "event tracked",
		Data:    gin.H{"event_id": events_id},
	})
}
