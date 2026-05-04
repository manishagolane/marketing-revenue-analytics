-- name: CreateCampaign :one
INSERT INTO campaigns (id, name, description, created_by, status, channel, budget, is_public, starts_at, ends_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetCampaignByID :one
SELECT * FROM campaigns
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetPublicCampaignByID :one
SELECT id, name, description, channel, status, is_public, starts_at, ends_at, created_at
FROM campaigns
WHERE id = $1
  AND is_public = TRUE
  AND deleted_at IS NULL;

-- name: UpdateCampaign :one
UPDATE campaigns
SET
  name        = $2,
  description = $3,
  channel     = $4,
  budget      = $5,
  is_public   = $6,
  starts_at   = $7,
  ends_at     = $8,
  updated_at  = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: UpdateCampaignStatus :one
UPDATE campaigns
SET status = $2, updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: DeleteCampaign :exec
UPDATE campaigns
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ListPublicCampaigns :many
SELECT id, name, description, channel, status, starts_at, ends_at, created_at
FROM campaigns
WHERE is_public = TRUE
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT  $1
OFFSET $2;

-- name: ListCampaigns :many
SELECT
  id, name, description, created_by,
  status, channel, budget, spend, revenue,
  is_public, starts_at, ends_at,
  created_at, updated_at
FROM campaigns
WHERE deleted_at IS NULL
  AND (sqlc.narg(status)::VARCHAR     IS NULL OR status     = sqlc.narg(status))
  AND (sqlc.narg(channel)::VARCHAR    IS NULL OR channel    = sqlc.narg(channel))
  AND (sqlc.narg(created_by)::VARCHAR IS NULL OR created_by = sqlc.narg(created_by))
  AND (sqlc.narg(is_public)::BOOLEAN  IS NULL OR is_public  = sqlc.narg(is_public))
  AND (sqlc.narg(from_date)::TIMESTAMPTZ IS NULL OR created_at >= sqlc.narg(from_date))
  AND (sqlc.narg(to_date)::TIMESTAMPTZ   IS NULL OR created_at <= sqlc.narg(to_date))
ORDER BY created_at DESC
LIMIT  sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: SearchCampaigns :many
SELECT * FROM campaigns
WHERE deleted_at IS NULL
  AND (sqlc.narg(created_by)::VARCHAR IS NULL OR created_by = sqlc.narg(created_by))
  AND to_tsvector('english', name || ' ' || COALESCE(description, ''))
      @@ plainto_tsquery('english', sqlc.arg(query))
ORDER BY created_at DESC
LIMIT  sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);