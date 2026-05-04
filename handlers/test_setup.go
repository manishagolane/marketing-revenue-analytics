package handlers

import (
	"context"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"marketing-revenue-analytics/models"
)

// ERROR MOCK

type errorString string

func (e errorString) Error() string { return string(e) }

// AUTH MOCK

type mockAuthStore struct{}

func (m *mockAuthStore) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	switch email {
	case "exists@test.com":
		hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

		return models.User{
			ID:       "1",
			Email:    email,
			Password: string(hash),
			Role:     "marketer",
		}, nil
	default:
		return models.User{}, errorString("not found")
	}
}

func (m *mockAuthStore) GetUserByID(ctx context.Context, id string) (models.User, error) {
	return models.User{ID: id, Email: "test@test.com"}, nil
}

func (m *mockAuthStore) CreateUser(ctx context.Context, params models.CreateUserParams) (models.User, error) {
	return models.User{ID: "1", Email: params.Email, Role: params.Role}, nil
}

func (m *mockAuthStore) UpdateUserProfile(ctx context.Context, params models.UpdateUserProfileParams) (models.User, error) {
	return models.User{ID: params.ID, Name: params.Name}, nil
}

// CAMPAIGN MOCK

type mockCampaignStore struct{}

func (m *mockCampaignStore) CreateCampaign(ctx context.Context, p models.CreateCampaignParams) (models.Campaign, error) {
	return models.Campaign{ID: "c1", Name: p.Name}, nil
}

func (m *mockCampaignStore) GetCampaignByID(ctx context.Context, id string) (models.Campaign, error) {
	return models.Campaign{ID: id, Status: "active", CreatedBy: "u1"}, nil
}

func (m *mockCampaignStore) GetPublicCampaignByID(ctx context.Context, id string) (models.GetPublicCampaignByIDRow, error) {
	return models.GetPublicCampaignByIDRow{}, nil
}

func (m *mockCampaignStore) UpdateCampaign(ctx context.Context, p models.UpdateCampaignParams) (models.Campaign, error) {
	return models.Campaign{ID: p.ID}, nil
}

func (m *mockCampaignStore) UpdateCampaignStatus(ctx context.Context, p models.UpdateCampaignStatusParams) (models.Campaign, error) {
	return models.Campaign{ID: p.ID, Status: p.Status}, nil
}

func (m *mockCampaignStore) DeleteCampaign(ctx context.Context, id string) error {
	return nil
}

func (m *mockCampaignStore) ListCampaigns(ctx context.Context, p models.ListCampaignsParams) ([]models.ListCampaignsRow, error) {
	return []models.ListCampaignsRow{}, nil
}

func (m *mockCampaignStore) ListPublicCampaigns(ctx context.Context, p models.ListPublicCampaignsParams) ([]models.ListPublicCampaignsRow, error) {
	return []models.ListPublicCampaignsRow{}, nil
}

func (m *mockCampaignStore) SearchCampaigns(ctx context.Context, p models.SearchCampaignsParams) ([]models.Campaign, error) {
	return []models.Campaign{}, nil
}

// EVENT MOCK

type mockEventStore struct{}

func (m *mockEventStore) GetCampaignByID(ctx context.Context, id string) (models.Campaign, error) {
	return models.Campaign{ID: id, Status: "active"}, nil
}

// SQS MOCK (IMPORTANT FIX)

type mockQueue struct{}

func (m *mockQueue) SendSqsMessageWithBody(ctx context.Context, payload interface{}, queueURL string) error {
	return nil
}

// ANALYTICS MOCK

type mockAnalyticsStore struct{}

func (m *mockAnalyticsStore) GetDailyMetrics(ctx context.Context, p models.GetDailyMetricsParams) ([]models.GetDailyMetricsRow, error) {
	return []models.GetDailyMetricsRow{}, nil
}

func (m *mockAnalyticsStore) GetWeeklyMetrics(ctx context.Context, p models.GetWeeklyMetricsParams) ([]models.GetWeeklyMetricsRow, error) {
	return []models.GetWeeklyMetricsRow{}, nil
}

func (m *mockAnalyticsStore) GetMonthlyMetrics(ctx context.Context, p models.GetMonthlyMetricsParams) ([]models.GetMonthlyMetricsRow, error) {
	return []models.GetMonthlyMetricsRow{}, nil
}

func (m *mockAnalyticsStore) GetCampaignSummary(ctx context.Context, p models.GetCampaignSummaryParams) (models.GetCampaignSummaryRow, error) {
	return models.GetCampaignSummaryRow{}, nil
}

// ENGAGEMENT MOCK

type mockEngagementStore struct {
	getFunnelStatsFn func(ctx context.Context, p models.GetFunnelStatsParams) ([]models.GetFunnelStatsRow, error)
	getTimeSpentFn   func(ctx context.Context, p models.GetTimeSpentPerSessionParams) ([]models.GetTimeSpentPerSessionRow, error)
	getClickPathFn   func(ctx context.Context, p models.GetClickPathParams) ([]models.GetClickPathRow, error)
}

func (m *mockEngagementStore) GetFunnelStats(ctx context.Context, p models.GetFunnelStatsParams) ([]models.GetFunnelStatsRow, error) {
	if m.getFunnelStatsFn != nil {
		return m.getFunnelStatsFn(ctx, p)
	}
	return []models.GetFunnelStatsRow{}, nil
}

func (m *mockEngagementStore) GetTimeSpentPerSession(ctx context.Context, p models.GetTimeSpentPerSessionParams) ([]models.GetTimeSpentPerSessionRow, error) {
	if m.getTimeSpentFn != nil {
		return m.getTimeSpentFn(ctx, p)
	}
	return []models.GetTimeSpentPerSessionRow{}, nil
}

func (m *mockEngagementStore) GetClickPath(ctx context.Context, p models.GetClickPathParams) ([]models.GetClickPathRow, error) {
	if m.getClickPathFn != nil {
		return m.getClickPathFn(ctx, p)
	}
	return []models.GetClickPathRow{}, nil
}

type mockCache struct{}

func (m *mockCache) Get(ctx context.Context, key string) (string, error) {
	return "", errorString("cache miss")
}

func (m *mockCache) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	return nil
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockCache) GetBool(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (m *mockCache) SetBool(ctx context.Context, key string, value bool, expiration time.Duration) error {
	return nil
}

func setupAuthHandler() *AuthHandler {
	return &AuthHandler{
		queries:    &mockAuthStore{},
		cache:      &mockCache{},
		jwtManager: &mockJWT{},
	}
}

func setupCampaignHandler() *CampaignHandler {
	return &CampaignHandler{
		queries: &mockCampaignStore{},
		logger:  zap.NewNop(),
	}
}

func setupEventHandler() *EventHandler {
	return &EventHandler{
		queries: &mockEventStore{},
		logger:  zap.NewNop(),
		clients: &mockQueue{},
	}
}

func setupAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{
		queries: &mockAnalyticsStore{},
		cache:   &mockCache{}, //
	}
}

func setupEngagementHandler() *EngagementHandler {
	return &EngagementHandler{
		queries: &mockEngagementStore{},
	}
}

type mockJWT struct{}

func (m *mockJWT) GenerateToken(userID, email, role, tokenType string, ttl time.Duration) (string, error) {
	return "header.payload.signature", nil
}
