package handlers

import (
	"context"
	"time"

	"marketing-revenue-analytics/models"
)

type AuthStore interface {
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetUserByID(ctx context.Context, id string) (models.User, error)
	CreateUser(ctx context.Context, params models.CreateUserParams) (models.User, error)
	UpdateUserProfile(ctx context.Context, params models.UpdateUserProfileParams) (models.User, error)
}

type CampaignStore interface {
	CreateCampaign(ctx context.Context, p models.CreateCampaignParams) (models.Campaign, error)
	GetCampaignByID(ctx context.Context, id string) (models.Campaign, error)
	GetPublicCampaignByID(ctx context.Context, id string) (models.GetPublicCampaignByIDRow, error)
	UpdateCampaign(ctx context.Context, p models.UpdateCampaignParams) (models.Campaign, error)
	UpdateCampaignStatus(ctx context.Context, p models.UpdateCampaignStatusParams) (models.Campaign, error)
	DeleteCampaign(ctx context.Context, id string) error
	ListCampaigns(ctx context.Context, p models.ListCampaignsParams) ([]models.ListCampaignsRow, error)
	ListPublicCampaigns(ctx context.Context, p models.ListPublicCampaignsParams) ([]models.ListPublicCampaignsRow, error)
	SearchCampaigns(ctx context.Context, p models.SearchCampaignsParams) ([]models.Campaign, error)
}

type EventStore interface {
	GetCampaignByID(ctx context.Context, id string) (models.Campaign, error)
}

type AnalyticsStore interface {
	GetDailyMetrics(ctx context.Context, p models.GetDailyMetricsParams) ([]models.GetDailyMetricsRow, error)
	GetWeeklyMetrics(ctx context.Context, p models.GetWeeklyMetricsParams) ([]models.GetWeeklyMetricsRow, error)
	GetMonthlyMetrics(ctx context.Context, p models.GetMonthlyMetricsParams) ([]models.GetMonthlyMetricsRow, error)
	GetCampaignSummary(ctx context.Context, p models.GetCampaignSummaryParams) (models.GetCampaignSummaryRow, error)
}

type EngagementStore interface {
	GetFunnelStats(ctx context.Context, p models.GetFunnelStatsParams) ([]models.GetFunnelStatsRow, error)
	GetTimeSpentPerSession(ctx context.Context, p models.GetTimeSpentPerSessionParams) ([]models.GetTimeSpentPerSessionRow, error)
	GetClickPath(ctx context.Context, p models.GetClickPathParams) ([]models.GetClickPathRow, error)
}

type CacheStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	GetBool(ctx context.Context, key string) (bool, error)
	SetBool(ctx context.Context, key string, value bool, expiration time.Duration) error
}

type QueueStore interface {
	SendSqsMessageWithBody(ctx context.Context, payload interface{}, queueURL string) error
}

type TokenService interface {
	GenerateToken(userID, email, role, tokenType string, ttl time.Duration) (string, error)
}

var (
	_ AuthStore       = (*models.Queries)(nil)
	_ CampaignStore   = (*models.Queries)(nil)
	_ EventStore      = (*models.Queries)(nil)
	_ AnalyticsStore  = (*models.Queries)(nil)
	_ EngagementStore = (*models.Queries)(nil)
)
