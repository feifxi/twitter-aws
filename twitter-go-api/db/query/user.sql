-- name: CreateUser :one
INSERT INTO users (
  username, email, display_name, bio, avatar_url, role, provider
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetUser :one
SELECT sqlc.embed(u),
  EXISTS(SELECT 1 FROM follows f WHERE f.following_id = u.id AND f.follower_id = sqlc.narg('viewer_id')) AS is_following
FROM users u
WHERE u.id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: UpdateUserProfile :one
UPDATE users
SET
  bio = $2,
  display_name = $3,
  avatar_url = $4,
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: SearchUsers :many
SELECT sqlc.embed(u),
  EXISTS(SELECT 1 FROM follows f WHERE f.following_id = u.id AND f.follower_id = sqlc.narg('viewer_id')) AS is_following
FROM users u
WHERE u.username ILIKE '%' || $1 || '%'
   OR u.display_name ILIKE '%' || $1 || '%'
ORDER BY u.followers_count DESC
LIMIT $2 OFFSET $3;

-- name: CountSearchUsers :one
SELECT COUNT(*)
FROM users u
WHERE u.username ILIKE '%' || $1 || '%'
   OR u.display_name ILIKE '%' || $1 || '%';

-- name: ListFollowersUsers :many
SELECT sqlc.embed(u),
  EXISTS(SELECT 1 FROM follows f2 WHERE f2.following_id = u.id AND f2.follower_id = sqlc.narg('viewer_id')) AS is_following
FROM users u
JOIN follows f ON u.id = f.follower_id
WHERE f.following_id = $1
ORDER BY f.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountFollowersUsers :one
SELECT COUNT(*)
FROM follows f
WHERE f.following_id = $1;

-- name: ListFollowingUsers :many
SELECT sqlc.embed(u),
  EXISTS(SELECT 1 FROM follows f2 WHERE f2.following_id = u.id AND f2.follower_id = sqlc.narg('viewer_id')) AS is_following
FROM users u
JOIN follows f ON u.id = f.following_id
WHERE f.follower_id = $1
ORDER BY f.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountFollowingUsers :one
SELECT COUNT(*)
FROM follows f
WHERE f.follower_id = $1;

-- name: ListSuggestedUsers :many
SELECT sqlc.embed(u),
  EXISTS(SELECT 1 FROM follows f2 WHERE f2.following_id = u.id AND f2.follower_id = sqlc.narg('viewer_id')) AS is_following
FROM users u
LEFT JOIN follows f ON f.following_id = u.id AND f.follower_id = $1
WHERE u.id != $1
ORDER BY (CASE WHEN f.follower_id IS NULL THEN 0 ELSE 1 END) ASC, u.followers_count DESC
LIMIT $2 OFFSET $3;

-- name: CountSuggestedUsers :one
SELECT COUNT(*)
FROM users u
WHERE u.id != $1;

-- name: ListTopUsers :many
SELECT * FROM users
ORDER BY followers_count DESC
LIMIT $1 OFFSET $2;

-- name: CountTopUsers :one
SELECT COUNT(*) FROM users;

-- name: IsFollowing :one
SELECT EXISTS(
  SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2
);

-- name: GetUsersByIDs :many
SELECT sqlc.embed(u),
  EXISTS(SELECT 1 FROM follows f WHERE f.following_id = u.id AND f.follower_id = sqlc.narg('viewer_id')) AS is_following
FROM users u
WHERE u.id = ANY(@user_ids::bigint[]);
