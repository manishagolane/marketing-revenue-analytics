-- name: UpsertDailyMetrics :exec
INSERT INTO campaign_daily_metrics (
  campaign_id, date, impressions, clicks, conversions
)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (campaign_id, date)
DO UPDATE SET
  impressions = campaign_daily_metrics.impressions + EXCLUDED.impressions,
  clicks      = campaign_daily_metrics.clicks + EXCLUDED.clicks,
  conversions = campaign_daily_metrics.conversions + EXCLUDED.conversions;