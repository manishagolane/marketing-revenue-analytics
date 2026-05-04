package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"

	"marketing-revenue-analytics/internal/dto"
	"marketing-revenue-analytics/models"
	"marketing-revenue-analytics/utils"
)

type AnalyticsHandler struct {
	queries AnalyticsStore
	cache   CacheStore
}

func NewAnalyticsHandler(q AnalyticsStore, cache CacheStore) *AnalyticsHandler {
	return &AnalyticsHandler{
		queries: q,
		cache:   cache,
	}
}

// GET /api/v1/analytics/daily?campaign_id=&channel=&from=&to=&page=&limit=
func (h *AnalyticsHandler) Daily(c *gin.Context) {
	params, err := parseAnalyticsQuery(c)
	if err != nil {
		SendBadRequestError(c, err)
		return
	}

	ctx := c.Request.Context()

	cacheKey := fmt.Sprintf(
		"analytics:daily:%s:%s:%s:%d:%d",
		params.CampaignID,
		params.From,
		params.To,
		params.Page,
		params.Limit,
	)

	cached, err := h.cache.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		c.Data(http.StatusOK, "application/json", []byte(cached))
		return
	}

	rows, err := h.queries.GetDailyMetrics(ctx, models.GetDailyMetricsParams{
		CampaignID: utils.ToNullString(params.CampaignID),
		Channel:    utils.ToNullString(params.Channel),
		FromDate:   utils.ToNullDate(params.From),
		ToDate:     utils.ToNullDate(params.To),
		PageLimit:  int32(params.Limit),
		PageOffset: int32((params.Page - 1) * params.Limit),
	})
	if err != nil {
		SendApplicationError(c, err)
		return
	}

	response := dto.APIResponse{
		Status:  "success",
		Message: "daily metrics",
		Data: gin.H{
			"metrics": rows,
			"page":    params.Page,
			"limit":   params.Limit,
		},
	}

	jsonData, _ := json.Marshal(response)
	_ = h.cache.Set(ctx, cacheKey, string(jsonData), 5*time.Minute)

	c.JSON(http.StatusOK, response)
}

// GET /api/v1/analytics/weekly
func (h *AnalyticsHandler) Weekly(c *gin.Context) {
	params, err := parseAnalyticsQuery(c)
	if err != nil {
		SendBadRequestError(c, err)
		return
	}

	rows, err := h.queries.GetWeeklyMetrics(c.Request.Context(), models.GetWeeklyMetricsParams{
		CampaignID: utils.ToNullString(params.CampaignID),
		Channel:    utils.ToNullString(params.Channel),
		FromDate:   utils.ToNullDate(params.From),
		ToDate:     utils.ToNullDate(params.To),
		PageLimit:  int32(params.Limit),
		PageOffset: int32((params.Page - 1) * params.Limit),
	})
	if err != nil {
		SendApplicationError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  "success",
		Message: "weekly metrics",
		Data: gin.H{
			"metrics": rows,
			"page":    params.Page,
			"limit":   params.Limit,
		},
	})
}

// GET /api/v1/analytics/monthly
func (h *AnalyticsHandler) Monthly(c *gin.Context) {
	params, err := parseAnalyticsQuery(c)
	if err != nil {
		SendBadRequestError(c, err)
		return
	}

	rows, err := h.queries.GetMonthlyMetrics(c.Request.Context(), models.GetMonthlyMetricsParams{
		CampaignID: utils.ToNullString(params.CampaignID),
		Channel:    utils.ToNullString(params.Channel),
		FromDate:   utils.ToNullDate(params.From),
		ToDate:     utils.ToNullDate(params.To),
		PageLimit:  int32(params.Limit),
		PageOffset: int32((params.Page - 1) * params.Limit),
	})
	if err != nil {
		SendApplicationError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  "success",
		Message: "monthly metrics",
		Data: gin.H{
			"metrics": rows,
			"page":    params.Page,
			"limit":   params.Limit,
		},
	})
}

// GET /api/v1/analytics/campaigns/:id/summary
// Public — no JWT, anonymized (Part 1 unauthenticated access requirement)
func (h *AnalyticsHandler) PublicSummary(c *gin.Context) {

	var q dto.PublicSummaryQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		SendBadRequestError(c, err)
		return
	}

	if q.From != "" {
		if _, err := time.Parse("2006-01-02", q.From); err != nil {
			SendBadRequestError(c, errors.New("invalid 'from' date"))
			return
		}
	}
	if q.To != "" {
		if _, err := time.Parse("2006-01-02", q.To); err != nil {
			SendBadRequestError(c, errors.New("invalid 'to' date"))
			return
		}
	}

	row, err := h.queries.GetCampaignSummary(c.Request.Context(), models.GetCampaignSummaryParams{
		ID:       c.Param("id"),
		FromDate: utils.ToNullDate(q.From),
		ToDate:   utils.ToNullDate(q.To),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			SendNotFoundError(c, errors.New("campaign not found or not public"))
			return
		}
		SendApplicationError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  "success",
		Message: "campaign summary",
		Data:    gin.H{"summary": row},
	})
}

type analyticsParams struct {
	CampaignID string
	Channel    string
	From       string
	To         string
	Page       int
	Limit      int
}

func parseAnalyticsQuery(c *gin.Context) (*analyticsParams, error) {
	var q dto.AnalyticsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		return nil, err
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 30
	}

	// Validate date format upfront — clear error message
	if q.From != "" {
		if _, err := time.Parse("2006-01-02", q.From); err != nil {
			return nil, errors.New("invalid 'from' date, use YYYY-MM-DD")
		}
	}
	if q.To != "" {
		if _, err := time.Parse("2006-01-02", q.To); err != nil {
			return nil, errors.New("invalid 'to' date, use YYYY-MM-DD")
		}
	}

	return &analyticsParams{
		CampaignID: q.CampaignID,
		Channel:    q.Channel,
		From:       q.From,
		To:         q.To,
		Page:       q.Page,
		Limit:      q.Limit,
	}, nil
}
