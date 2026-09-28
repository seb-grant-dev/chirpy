-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1
)
RETURNING *;

-- name: GetUserFromEmail :one
SELECT
  *
FROM
  users 
WHERE
  users.email = $1;



-- name: ResetUsers :exec
DELETE FROM users;
