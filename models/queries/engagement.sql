-- name: GetFunnelStats :many
-- Drop-off rates: count unique sessions at each funnel stage
SELECT
  step,
  COUNT(DISTINCT session_id)           AS sessions,
  COUNT(*)                             AS total_events
FROM event_logs
WHERE campaign_id = $1
  AND step        IS NOT NULL
  AND session_id  IS NOT NULL
  AND (sqlc.narg(from_date)::TIMESTAMPTZ IS NULL OR occurred_at >= sqlc.narg(from_date))
  AND (sqlc.narg(to_date)::TIMESTAMPTZ   IS NULL OR occurred_at <= sqlc.narg(to_date))
GROUP BY step
ORDER BY
  CASE step
    WHEN 'ad'       THEN 1
    WHEN 'landing'  THEN 2
    WHEN 'signup'   THEN 3
    WHEN 'purchase' THEN 4
    ELSE 5
  END;

-- name: GetTimeSpentPerSession :many
-- Time spent: difference between first and last event per session
SELECT
  session_id,
  MIN(occurred_at)                                          AS session_start,
  MAX(occurred_at)                                          AS session_end,
  EXTRACT(EPOCH FROM (MAX(occurred_at) - MIN(occurred_at)))::DOUBLE PRECISION AS duration_seconds,
  COUNT(*)                                                  AS event_count
FROM event_logs
WHERE campaign_id  = $1
  AND session_id   IS NOT NULL
  AND (sqlc.narg(from_date)::TIMESTAMPTZ IS NULL OR occurred_at >= sqlc.narg(from_date))
  AND (sqlc.narg(to_date)::TIMESTAMPTZ   IS NULL OR occurred_at <= sqlc.narg(to_date))
GROUP BY session_id
ORDER BY session_start DESC
LIMIT  sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: GetClickPath :many
-- Which steps users moved through, in order, per session
SELECT
  session_id,
  ARRAY_AGG(step ORDER BY occurred_at) AS path
FROM event_logs
WHERE campaign_id = $1
  AND step        IS NOT NULL
  AND session_id  IS NOT NULL
  AND (sqlc.narg(from_date)::TIMESTAMPTZ IS NULL OR occurred_at >= sqlc.narg(from_date))
  AND (sqlc.narg(to_date)::TIMESTAMPTZ   IS NULL OR occurred_at <= sqlc.narg(to_date))
GROUP BY session_id
LIMIT  sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);