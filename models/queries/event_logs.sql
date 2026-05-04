-- name: CreateEventLog :one
INSERT INTO event_logs (
  id, campaign_id, event_type,
  source_url, ip_address, user_agent,
  metadata, occurred_at, session_id, step
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetEventCountsByCampaign :one
SELECT
  campaign_id,
  COUNT(*) FILTER (WHERE event_type = 'impression') AS impressions,
  COUNT(*) FILTER (WHERE event_type = 'click')      AS clicks,
  COUNT(*) FILTER (WHERE event_type = 'conversion') AS conversions
FROM event_logs
WHERE campaign_id = $1
  AND (sqlc.narg(from_date)::TIMESTAMPTZ IS NULL OR occurred_at >= sqlc.narg(from_date))
  AND (sqlc.narg(to_date)::TIMESTAMPTZ   IS NULL OR occurred_at <= sqlc.narg(to_date))
GROUP BY campaign_id;

-- name: GetEventsByType :many
SELECT * FROM event_logs
WHERE campaign_id = $1
  AND event_type  = $2
  AND (sqlc.narg(from_date)::TIMESTAMPTZ IS NULL OR occurred_at >= sqlc.narg(from_date))
  AND (sqlc.narg(to_date)::TIMESTAMPTZ   IS NULL OR occurred_at <= sqlc.narg(to_date))
ORDER BY occurred_at DESC
LIMIT  sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

SELECT step, COUNT(DISTINCT session_id)
FROM event_logs
WHERE campaign_id = $1
GROUP BY step
ORDER BY step;

SELECT session_id, MIN(occurred_at), MAX(occurred_at)
FROM event_logs
WHERE campaign_id = $1
GROUP BY session_id;