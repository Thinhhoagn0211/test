-- name: CreateUser :one
INSERT INTO users (
  email,
  username,
  password,
  password_hash,
  phone,
  fullname,
  avatar,
  state,
  role,
  created_at,
  update_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;


-- name: GetUsersAsc :many
SELECT * FROM users
WHERE (username ILIKE '%' || $1 || '%' OR fullname ILIKE '%' || $1 || '%' OR $1 IS NULL)
  AND state = $2
ORDER BY 
  CASE WHEN $3 = 'id' THEN id END ASC,
  CASE WHEN $3 = 'username' THEN username END ASC,
  CASE WHEN $3 = 'fullname' THEN fullname END ASC
LIMIT $4 OFFSET $5;

-- name: GetUsersDesc :many
SELECT * FROM users
WHERE (username ILIKE '%' || $1 || '%' OR fullname ILIKE '%' || $1 || '%' OR $1 IS NULL)
  AND state = $2
ORDER BY
  CASE WHEN $3 = 'id' THEN id END DESC,
  CASE WHEN $3 = 'username' THEN username END DESC,
  CASE WHEN $3 = 'fullname' THEN fullname END DESC
LIMIT $4 OFFSET $5;


-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1;

-- name: UpdateUser :one
UPDATE users
SET
  fullname = COALESCE(sqlc.narg(fullname), fullname),
  email = COALESCE(sqlc.narg(email), email),
  phone = COALESCE(sqlc.narg(phone), phone),
  password = COALESCE(sqlc.narg(password), password),
  avatar = COALESCE(sqlc.narg(avatar), avatar),
  role = COALESCE(sqlc.narg(role), role),
  password_hash = COALESCE(sqlc.narg(password_hash), password_hash)
WHERE
  username = sqlc.arg(username)
RETURNING *;

-- name: DeleteUser :one
UPDATE users
SET state = $1
WHERE id = $2
RETURNING *;
