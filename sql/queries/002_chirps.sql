-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: GetAllUserChirp :many
SELECT  * 
FROM chirps 
WHERE user_id = $1;

-- name: GetAllChirps :many
SELECT  *
FROM chirps;

-- name: GetAllUserChirpASC :many
SELECT  * 
FROM chirps 
WHERE user_id = $1
ORDER BY created_at ASC;

-- name: GetAllChirpsASC :many
SELECT  *
FROM chirps
ORDER BY created_at ASC;

-- name: GetAllUserChirpDESC :many
SELECT  * 
FROM chirps 
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetAllChirpsDESC :many
SELECT  *
FROM chirps
ORDER BY created_at DESC;

-- name: GetOneChirp :one
SELECT  *
FROM chirps
WHERE id = $1;

-- name: RemoveChirp :one
DELETE From chirps
WHERE id = $1 AND user_id = $2
RETURNING *;
