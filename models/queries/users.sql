-- name: CreateUser :one
INSERT INTO users (id, name, email, password, phone, role)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUserProfile :one
UPDATE users
SET
  name       = $2,
  phone      = $3,
  bio        = $4,
  picture    = $5,
  updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;