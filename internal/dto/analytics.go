package dto

type AnalyticsQuery struct {
	CampaignID string `form:"campaign_id"`
	Channel    string `form:"channel"`
	From       string `form:"from"` // YYYY-MM-DD
	To         string `form:"to"`   // YYYY-MM-DD
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=30"`
}

type PublicSummaryQuery struct {
	From string `form:"from"` // YYYY-MM-DD
	To   string `form:"to"`   // YYYY-MM-DD
}

type DailyMetricResponse struct {
	Date    string  `json:"date"`
	Imps    int64   `json:"impressions"`
	Clicks  int64   `json:"clicks"`
	Revenue float64 `json:"revenue"`
}

type AnalyticsResponse struct {
	Metrics []DailyMetricResponse `json:"metrics"`
	Page    int                   `json:"page"`
	Limit   int                   `json:"limit"`
}
