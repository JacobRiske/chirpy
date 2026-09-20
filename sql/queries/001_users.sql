-- name: CreateUser :one
Insert into users (id, created_at, updated_at, email)
Values(
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1
)
Returning *;

-- name: ResetUsers :exec
TRUNCATE TABLE chirps, users RESTART IDENTITY CASCADE;

-- name: LookUpUser :one
Select id From users Where email = $1;