package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"marketing-revenue-analytics/internal/dto"
	"marketing-revenue-analytics/models"
	"marketing-revenue-analytics/utils"
)

type EngagementHandler struct {
	queries EngagementStore
}

func NewEngagementHandler(q EngagementStore) *EngagementHandler {
	return &EngagementHandler{queries: q}
}

// GET /api/v1/engagement/:campaign_id/funnel
// Shows how many unique sessions reached each stage
// Drop-off = sessions at stage N minus sessions at stage N+1
func (h *EngagementHandler) Funnel(c *gin.Context) {
	campaignID := c.Param("campaign_id")
	q, err := parseEngagementQuery(c)
	if err != nil {
		SendBadRequestError(c, err)
		return
	}

	rows, err := h.queries.GetFunnelStats(c.Request.Context(), models.GetFunnelStatsParams{
		CampaignID: campaignID,
		FromDate:   utils.ToNullTime(q.from),
		ToDate:     utils.ToNullTime(q.to),
	})
	if err != nil {
		SendApplicationError(c, err)
		return
	}

	// Calculate drop-off rates between stages
	type FunnelStage struct {
		Step        string  `json:"step"`
		Sessions    int64   `json:"sessions"`
		DropOffRate float64 `json:"drop_off_rate"` // % who did NOT proceed to next stage
	}

	stages := make([]FunnelStage, 0, len(rows))
	for i, row := range rows {
		stage := FunnelStage{
			Step:     row.Step.String,
			Sessions: row.Sessions,
		}
		if i > 0 && rows[i-1].Sessions > 0 {
			prev := rows[i-1].Sessions
			dropped := prev - row.Sessions
			stage.DropOffRate = float64(dropped) / float64(prev) * 100
		}
		stages = append(stages, stage)
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  "success",
		Message: "funnel stats",
		Data:    gin.H{"funnel": stages},
	})
}

// GET /api/v1/engagement/:campaign_id/time-spent
// Average and per-session time spent on campaign pages
func (h *EngagementHandler) TimeSpent(c *gin.Context) {
	campaignID := c.Param("campaign_id")
	q, err := parseEngagementQuery(c)
	if err != nil {
		SendBadRequestError(c, err)
		return
	}

	rows, err := h.queries.GetTimeSpentPerSession(c.Request.Context(), models.GetTimeSpentPerSessionParams{
		CampaignID: campaignID,
		FromDate:   utils.ToNullTime(q.from),
		ToDate:     utils.ToNullTime(q.to),
		PageLimit:  int32(q.Limit),
		PageOffset: int32((q.Page - 1) * q.Limit),
	})
	if err != nil {
		SendApplicationError(c, err)
		return
	}

	// Compute average duration across sessions
	var totalSeconds float64
	for _, r := range rows {
		totalSeconds += r.DurationSeconds
	}

	var avgSeconds float64
	if len(rows) > 0 {
		avgSeconds = totalSeconds / float64(len(rows))
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  "success",
		Message: "time spent",
		Data: gin.H{
			"sessions":             rows,
			"avg_duration_seconds": avgSeconds,
			"page":                 q.Page,
			"limit":                q.Limit,
		},
	})
}

// GET /api/v1/engagement/:campaign_id/click-path
// Shows the sequence of steps each session went through
func (h *EngagementHandler) ClickPath(c *gin.Context) {
	campaignID := c.Param("campaign_id")
	q, err := parseEngagementQuery(c)
	if err != nil {
		SendBadRequestError(c, err)
		return
	}

	rows, err := h.queries.GetClickPath(c.Request.Context(), models.GetClickPathParams{
		CampaignID: campaignID,
		FromDate:   utils.ToNullTime(q.from),
		ToDate:     utils.ToNullTime(q.to),
		PageLimit:  int32(q.Limit),
		PageOffset: int32((q.Page - 1) * q.Limit),
	})
	if err != nil {
		SendApplicationError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  "success",
		Message: "click paths",
		Data: gin.H{
			"paths": rows,
			"page":  q.Page,
			"limit": q.Limit,
		},
	})
}

type engagementParams struct {
	from  *time.Time
	to    *time.Time
	Page  int
	Limit int
}

func parseEngagementQuery(c *gin.Context) (*engagementParams, error) {
	var q dto.EngagementQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		return nil, err
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 50
	}

	params := &engagementParams{Page: q.Page, Limit: q.Limit}

	if q.From != "" {
		t, err := time.Parse(time.RFC3339, q.From)
		if err != nil {
			return nil, errors.New("invalid 'from' date, use RFC3339")
		}
		params.from = &t
	}
	if q.To != "" {
		t, err := time.Parse(time.RFC3339, q.To)
		if err != nil {
			return nil, errors.New("invalid 'to' date, use RFC3339")
		}
		params.to = &t
	}

	return params, nil
}
