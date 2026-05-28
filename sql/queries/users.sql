-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT id, created_at, updated_at, email, hashed_password, is_chirpy_red
FROM users
WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET updated_at = NOW(),
    email = COALESCE($2, email),
    hashed_password = COALESCE($3, hashed_password)
WHERE id = $1
RETURNING *;

-- name: UpgradeUserToChirpyRed :one
UPDATE users
SET updated_at = NOW(),
    is_chirpy_red = TRUE
WHERE id = $1
RETURNING *;

-- name: GetChirpyRedStaus :one
SELECT is_chirpy_red
FROM users
WHERE id = $1;