-- name: CreateUserRefeashToken :one
Insert into refresh_tokens (id, token, created_at, updated_at, user_id, expires_at, revoked_at)
Values(
    gen_random_uuid(),
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    $4
)
Returning *;

-- name: ResetUsersRefeashToken :exec
TRUNCATE TABLE refresh_tokens, users RESTART IDENTITY CASCADE;

-- name: LookUpUserRefeashToken :one
Select token, revoked_at From refresh_tokens Where user_id = $1;

-- name: CheckToken :one
Select user_id, expires_at, revoked_at From refresh_tokens Where token = $1;

-- name: RevokedToken :one
UPDATE refresh_tokens
SET updated_at = NOW(),
    revoked_at = NOW()
WHERE token = $1
Returning *;

-- name: ResetRefeashToken :one
UPDATE refresh_tokens
SET token = $1,
    updated_at = NOW(),
    revoked_at = $2
WHERE user_id =$3
Returning *;
