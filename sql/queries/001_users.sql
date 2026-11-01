-- name: CreateUser :one
Insert into users (id, created_at, updated_at, email, pass)
Values(
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
Returning *;

-- name: ResetUsers :exec
TRUNCATE TABLE chirps, users RESTART IDENTITY CASCADE;

-- name: LookUpUserPass :one
Select * From users Where email = $1;