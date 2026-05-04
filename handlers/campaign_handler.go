package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	ulid "github.com/oklog/ulid/v2"
	"go.uber.org/zap"

	"marketing-revenue-analytics/internal/dto"
	"marketing-revenue-analytics/models"
	"marketing-revenue-analytics/utils"
)

type CampaignHandler struct {
	queries CampaignStore
	logger  *zap.Logger
}

func NewCampaignHandler(q CampaignStore, logger *zap.Logger) *CampaignHandler {
	return &CampaignHandler{
		queries: q,
		logger:  logger,
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// POST /api/v1/campaigns
func (h *CampaignHandler) Create(c *gin.Context) {
	var req dto.CreateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendBadRequestError(c, err)
		return
	}

	ctx := c.Request.Context()
	userID, err := utils.GetLoggedInUser(c)
	if err != nil {
		SendUnauthorizedError(c, err)
		return
	}

	startsAt, endsAt, err := parseDateRange(req.StartsAt, req.EndsAt)
	if err != nil {
		SendBadRequestError(c, err)
		return
	}

	campaign, err := h.queries.CreateCampaign(ctx, models.CreateCampaignParams{
		ID:          ulid.Make().String(),
		Name:        req.Name,
		Description: utils.ToNullString(req.Description),
		CreatedBy:   userID,
		Status:      "draft",
		Channel:     utils.ToNullString(req.Channel),
		Budget:      utils.ToNumeric(req.Budget),
		IsPublic:    req.IsPublic,
		StartsAt:    utils.ToNullTime(startsAt),
		EndsAt:      utils.ToNullTime(endsAt),
	})
	if err != nil {
		SendApplicationError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Status:  "success",
		Message: "campaign created",
		Data:    gin.H{"campaign": campaign},
	})
}

func (h *CampaignHandler) List(c *gin.Context) {
	var q dto.ListCampaignsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		SendBadRequestError(c, err)
		return
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 20
	}

	var from, to *time.Time
	if q.From != "" {
		t, err := time.Parse(time.RFC3339, q.From)
		if err != nil {
			SendBadRequestError(c, errors.New("invalid 'from' date, expected RFC3339"))
			return
		}
		from = &t
	}
	if q.To != "" {
		t, err := time.Parse(time.RFC3339, q.To)
		if err != nil {
			SendBadRequestError(c, errors.New("invalid 'to' date, expected RFC3339"))
			return
		}
		to = &t
	}

	userID := c.GetString("userID")
	role := c.GetString("roleID")

	// Role-aware visibility — never trust created_by from client
	// Admin → sees all campaigns (created_by = NULL → no filter)
	// Marketer → sees only their own campaigns
	// Analyst → sees all campaigns (read-only enforced by RBAC)
	var createdBy *string
	if role == "marketer" {
		createdBy = &userID
	}

	ctx := c.Request.Context()
	offset := (q.Page - 1) * q.Limit

	campaigns, err := h.queries.ListCampaigns(ctx, models.ListCampaignsParams{
		Status:     utils.ToNullString(q.Status),
		Channel:    utils.ToNullString(q.Channel),
		CreatedBy:  utils.ToNullStringPtr(createdBy),
		IsPublic:   utils.ToNullBool(q.IsPublic),
		FromDate:   utils.ToNullTime(from),
		ToDate:     utils.ToNullTime(to),
		PageLimit:  int32(q.Limit),
		PageOffset: int32(offset),
	})
	if err != nil {
		SendApplicationError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  "success",
		Message: "campaigns fetched",
		Data: gin.H{
			"campaigns": campaigns,
			"page":      q.Page,
			"limit":     q.Limit,
		},
	})
}

// GET /api/v1/campaigns/search?q=keyword
func (h *CampaignHandler) Search(c *gin.Context) {
	var q dto.SearchQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		SendBadRequestError(c, err)
		return
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 20
	}

	userID := c.GetString("userID")
	role := c.GetString("roleID")

	// Same visibility rule as List
	var createdBy *string
	if role == "marketer" {
		createdBy = &userID
	}

	ctx := c.Request.Context()
	offset := (q.Page - 1) * q.Limit

	campaigns, err := h.queries.SearchCampaigns(ctx, models.SearchCampaignsParams{
		Query:      q.Q,
		CreatedBy:  utils.ToNullStringPtr(createdBy),
		PageLimit:  int32(q.Limit),
		PageOffset: int32(offset),
	})
	if err != nil {
		SendApplicationError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  "success",
		Message: "search results",
		Data: gin.H{
			"campaigns": campaigns,
			"page":      q.Page,
			"limit":     q.Limit,
		},
	})
}

// GET /api/v1/campaigns/:id
func (h *CampaignHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()

	campaign, err := h.queries.GetCampaignByID(ctx, c.Param("id"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			SendNotFoundError(c, errors.New("campaign not found"))
			return
		}
		SendApplicationError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  "success",
		Message: "campaign fetched",
		Data:    gin.H{"campaign": campaign},
	})
}

// PATCH /api/v1/campaigns/:id
func (h *CampaignHandler) Update(c *gin.Context) {
	var req dto.UpdateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendBadRequestError(c, err)
		return
	}

	ctx := c.Request.Context()
	userID, err := utils.GetLoggedInUser(c)
	if err != nil {
		SendUnauthorizedError(c, err)
		return
	}

	role, _ := c.Get("roleID")

	startsAt, endsAt, err := parseDateRange(req.StartsAt, req.EndsAt)
	if err != nil {
		SendBadRequestError(c, err)
		return
	}
	// STEP 1: fetch campaign first (ownership check)
	campaign, err := h.queries.GetCampaignByID(ctx, c.Param("id"))
	if err != nil {
		SendNotFoundError(c, errors.New("campaign not found"))
		return
	}

	// STEP 2: ownership / RBAC check
	isOwner := campaign.CreatedBy == userID
	isAdmin := role == "admin"

	h.logger.Info("role:", zap.Any("role:", role))

	if !isOwner && !isAdmin {
		SendForbiddenError(c, errors.New("not allowed to update this campaign"))
		return
	}

	// STEP 3: perform update
	updated, err := h.queries.UpdateCampaign(ctx, models.UpdateCampaignParams{
		ID:          c.Param("id"),
		Name:        req.Name,
		Description: utils.ToNullString(strVal(req.Description)),
		Channel:     utils.ToNullString(req.Channel),
		Budget:      utils.ToNumeric(req.Budget),
		IsPublic:    req.IsPublic,
		StartsAt:    utils.ToNullTime(startsAt),
		EndsAt:      utils.ToNullTime(endsAt),
	})
	if err != nil {
		SendApplicationError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  "success",
		Message: "campaign updated",
		Data:    gin.H{"campaign": updated},
	})
}

// PATCH /api/v1/campaigns/:id/status
func (h *CampaignHandler) UpdateStatus(c *gin.Context) {
	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendBadRequestError(c, err)
		return
	}

	ctx := c.Request.Context()
	userID, _ := utils.GetLoggedInUser(c)
	role, _ := c.Get("roleID")

	campaign, err := h.queries.GetCampaignByID(ctx, c.Param("id"))
	if err != nil {
		SendNotFoundError(c, errors.New("campaign not found"))
		return
	}

	isOwner := campaign.CreatedBy == userID
	isAdmin := role == "admin"

	if !isOwner && !isAdmin {
		SendForbiddenError(c, errors.New("not allowed to update status"))
		return
	}

	updated, err := h.queries.UpdateCampaignStatus(ctx, models.UpdateCampaignStatusParams{
		ID:     campaign.ID,
		Status: req.Status,
	})
	if err != nil {
		SendApplicationError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  "success",
		Message: "status updated",
		Data:    gin.H{"campaign": updated},
	})
}

// DELETE /api/v1/campaigns/:id
func (h *CampaignHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := utils.GetLoggedInUser(c)
	if err != nil {
		SendUnauthorizedError(c, err)
		return
	}

	role, _ := c.Get("roleID")

	// fetch campaign first
	campaign, err := h.queries.GetCampaignByID(ctx, c.Param("id"))
	if err != nil {
		SendNotFoundError(c, errors.New("campaign not found"))
		return
	}

	isOwner := campaign.CreatedBy == userID
	isAdmin := role == "admin"

	if !isOwner && !isAdmin {
		SendForbiddenError(c, errors.New("not allowed to delete this campaign"))
		return
	}

	if err := h.queries.DeleteCampaign(ctx, campaign.ID); err != nil {
		SendApplicationError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  "success",
		Message: "campaign deleted",
	})
}

// GET /api/v1/campaigns/:id/preview  (no auth)
func (h *CampaignHandler) GetPublicPreview(c *gin.Context) {
	ctx := c.Request.Context()

	campaign, err := h.queries.GetPublicCampaignByID(ctx, c.Param("id"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			SendNotFoundError(c, errors.New("campaign not found"))
			return
		}
		SendApplicationError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  "success",
		Message: "campaign preview",
		Data:    gin.H{"campaign": campaign},
	})
}

// GET /api/v1/campaigns/public  (no auth)
func (h *CampaignHandler) ListPublic(c *gin.Context) {
	page := 1
	limit := 20

	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	ctx := c.Request.Context()
	offset := (page - 1) * limit

	campaigns, err := h.queries.ListPublicCampaigns(ctx, models.ListPublicCampaignsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		SendApplicationError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  "success",
		Message: "public campaigns",
		Data: gin.H{
			"campaigns": campaigns,
			"page":      page,
			"limit":     limit,
		},
	})
}

func parseDateRange(startsAt, endsAt *string) (*time.Time, *time.Time, error) {
	var s, e *time.Time
	if startsAt != nil && *startsAt != "" {
		t, err := time.Parse(time.RFC3339, *startsAt)
		if err != nil {
			return nil, nil, errors.New("invalid starts_at, use RFC3339")
		}
		s = &t
	}
	if endsAt != nil && *endsAt != "" {
		t, err := time.Parse(time.RFC3339, *endsAt)
		if err != nil {
			return nil, nil, errors.New("invalid ends_at, use RFC3339")
		}
		e = &t
	}
	if s != nil && e != nil && e.Before(*s) {
		return nil, nil, errors.New("ends_at must be after starts_at")
	}
	return s, e, nil
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
