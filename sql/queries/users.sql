-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING id, created_at, updated_at, email;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT id, created_at, updated_at, email, hashed_password
FROM users
WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET updated_at = NOW(),
    email = COALESCE($2, email),
    hashed_password = COALESCE($3, hashed_password)
WHERE id = $1
RETURNING id, created_at, updated_at, email;