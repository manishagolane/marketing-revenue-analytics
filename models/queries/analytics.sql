-- name: GetDailyMetrics :many
SELECT
  d.date,
  d.campaign_id,
  c.name                                                        AS campaign_name,
  c.channel,
  d.impressions,
  d.clicks,
  d.conversions,
  c.spend,
  c.revenue,
  CASE WHEN d.impressions > 0
    THEN ROUND((d.clicks::NUMERIC / d.impressions) * 100, 2)::DOUBLE PRECISION
    ELSE 0 END                                                  AS ctr,
  CASE WHEN d.clicks > 0
    THEN ROUND(c.spend / d.clicks, 2)::DOUBLE PRECISION
    ELSE 0 END                                                  AS cpc,
  CASE WHEN c.spend > 0
    THEN ROUND(((c.revenue - c.spend) / c.spend) * 100, 2)::DOUBLE PRECISION
    ELSE 0 END                                                  AS roi,
  CASE WHEN d.clicks > 0
    THEN ROUND((d.conversions::NUMERIC / d.clicks) * 100, 2)::DOUBLE PRECISION
    ELSE 0 END                                                  AS conversion_rate
FROM campaign_daily_metrics d
JOIN campaigns c ON c.id = d.campaign_id
WHERE c.deleted_at IS NULL
  AND (sqlc.narg(campaign_id)::VARCHAR IS NULL OR d.campaign_id = sqlc.narg(campaign_id))
  AND (sqlc.narg(channel)::VARCHAR     IS NULL OR c.channel     = sqlc.narg(channel))
  AND (sqlc.narg(from_date)::DATE      IS NULL OR d.date        >= sqlc.narg(from_date))
  AND (sqlc.narg(to_date)::DATE        IS NULL OR d.date        <= sqlc.narg(to_date))
ORDER BY d.date DESC
LIMIT  sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: GetWeeklyMetrics :many
SELECT
  DATE_TRUNC('week', d.date)::DATE                              AS week_start,
  d.campaign_id,
  c.name                                                        AS campaign_name,
  c.channel,
  SUM(d.impressions)                                            AS impressions,
  SUM(d.clicks)                                                 AS clicks,
  SUM(d.conversions)                                            AS conversions,
  c.spend,
  c.revenue,
  CASE WHEN SUM(d.impressions) > 0
    THEN ROUND((SUM(d.clicks)::NUMERIC / SUM(d.impressions)) * 100, 2)::DOUBLE PRECISION
    ELSE 0 END                                                  AS ctr,
  CASE WHEN SUM(d.clicks) > 0
    THEN ROUND(c.spend / SUM(d.clicks), 2)::DOUBLE PRECISION
    ELSE 0 END                                                  AS cpc,
  CASE WHEN c.spend > 0
    THEN ROUND(((c.revenue - c.spend) / c.spend) * 100, 2)::DOUBLE PRECISION
    ELSE 0 END                                                  AS roi,
  CASE WHEN SUM(d.clicks) > 0
    THEN ROUND((SUM(d.conversions)::NUMERIC / SUM(d.clicks)) * 100, 2)::DOUBLE PRECISION
    ELSE 0 END                                                  AS conversion_rate
FROM campaign_daily_metrics d
JOIN campaigns c ON c.id = d.campaign_id
WHERE c.deleted_at IS NULL
  AND (sqlc.narg(campaign_id)::VARCHAR IS NULL OR d.campaign_id = sqlc.narg(campaign_id))
  AND (sqlc.narg(channel)::VARCHAR     IS NULL OR c.channel     = sqlc.narg(channel))
  AND (sqlc.narg(from_date)::DATE      IS NULL OR d.date        >= sqlc.narg(from_date))
  AND (sqlc.narg(to_date)::DATE        IS NULL OR d.date        <= sqlc.narg(to_date))
GROUP BY week_start, d.campaign_id, c.name, c.channel, c.spend, c.revenue
ORDER BY week_start DESC
LIMIT  sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: GetMonthlyMetrics :many
SELECT
  DATE_TRUNC('month', d.date)::DATE                             AS month_start,
  d.campaign_id,
  c.name                                                        AS campaign_name,
  c.channel,
  SUM(d.impressions)                                            AS impressions,
  SUM(d.clicks)                                                 AS clicks,
  SUM(d.conversions)                                            AS conversions,
  c.spend,
  c.revenue,
  CASE WHEN SUM(d.impressions) > 0
    THEN ROUND((SUM(d.clicks)::NUMERIC / SUM(d.impressions)) * 100, 2)::DOUBLE PRECISION
    ELSE 0 END                                                  AS ctr,
  CASE WHEN SUM(d.clicks) > 0
    THEN ROUND(c.spend / SUM(d.clicks), 2)::DOUBLE PRECISION
    ELSE 0 END                                                  AS cpc,
  CASE WHEN c.spend > 0
    THEN ROUND(((c.revenue - c.spend) / c.spend) * 100, 2)::DOUBLE PRECISION
    ELSE 0 END                                                  AS roi,
  CASE WHEN SUM(d.clicks) > 0
    THEN ROUND((SUM(d.conversions)::NUMERIC / SUM(d.clicks)) * 100, 2)::DOUBLE PRECISION
    ELSE 0 END                                                  AS conversion_rate
FROM campaign_daily_metrics d
JOIN campaigns c ON c.id = d.campaign_id
WHERE c.deleted_at IS NULL
  AND (sqlc.narg(campaign_id)::VARCHAR IS NULL OR d.campaign_id = sqlc.narg(campaign_id))
  AND (sqlc.narg(channel)::VARCHAR     IS NULL OR c.channel     = sqlc.narg(channel))
  AND (sqlc.narg(from_date)::DATE      IS NULL OR d.date        >= sqlc.narg(from_date))
  AND (sqlc.narg(to_date)::DATE        IS NULL OR d.date        <= sqlc.narg(to_date))
GROUP BY month_start, d.campaign_id, c.name, c.channel, c.spend, c.revenue
ORDER BY month_start DESC
LIMIT  sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: GetCampaignSummary :one
SELECT
  c.id,
  c.name,
  c.channel,
  SUM(d.impressions)                                            AS total_impressions,
  SUM(d.clicks)                                                 AS total_clicks,
  SUM(d.conversions)                                            AS total_conversions,
  CASE WHEN SUM(d.impressions) > 0
    THEN ROUND((SUM(d.clicks)::NUMERIC / SUM(d.impressions)) * 100, 2)::DOUBLE PRECISION
    ELSE 0 END                                                  AS ctr
FROM campaign_daily_metrics d
JOIN campaigns c ON c.id = d.campaign_id
WHERE c.id = $1
  AND c.is_public    = TRUE
  AND c.deleted_at   IS NULL
  AND (sqlc.narg(from_date)::DATE IS NULL OR d.date >= sqlc.narg(from_date))
  AND (sqlc.narg(to_date)::DATE   IS NULL OR d.date <= sqlc.narg(to_date))
GROUP BY c.id, c.name, c.channel;