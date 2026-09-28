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

-- name: GetUserForEmail :one
SELECT
  *
FROM
  users 
WHERE
  users.email = $1;

-- name: GetUserForEmailAndPassword :one
SELECT
  *
FROM
  users
WHERE
  users.email = $1
  AND users.hashed_password = $2;



-- name: ResetUsers :exec
DELETE FROM users;
