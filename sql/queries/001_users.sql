-- name: CreateUser :one
Insert into users (id, created_at, updated_at, is_chirpy_red, email, pass)
Values(
    gen_random_uuid(),
    NOW(),
    NOW(),
    false,
    $1,
    $2
)
Returning *;

-- name: ResetUsers :exec
TRUNCATE TABLE chirps, users RESTART IDENTITY CASCADE;

-- name: LookUpUserPass :one
Select * From users Where email = $1;

-- name: UpdateUserData :one
UPDATE users
SET email = $1,
    pass = $2,
    updated_at = NOW()
WHERE id = $3
Returning *;

-- name: UpdateUserChirpyRed :one
UPDATE users
SET is_chirpy_red = $1
WHERE id = $2
Returning *;
