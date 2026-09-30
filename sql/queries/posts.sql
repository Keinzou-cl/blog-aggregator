-- name: CreatePost :one
INSERT INTO posts (id, feed_id, title, url, description, created_at, updated_at, published_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetPostsForUser :many
SELECT
    posts.*,
    feeds.name AS feed_name,
    users.name AS user_name
FROM 
    posts
    JOIN feeds ON posts.feed_id = feeds.id
    JOIN users ON feeds.user_id = users.id
WHERE
    users.name = $1
ORDER BY
    posts.published_at DESC
LIMIT
    $2; 