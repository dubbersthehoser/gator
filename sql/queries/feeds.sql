
-- name: CreateFeed :one
INSERT INTO feeds(
	id,
	created_at,
	updated_at,
	url,
	name,
	user_id
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: MarkFeedFetched :exec
UPDATE feeds 
SET last_fetched_at = NOW(), updated_at = NOW() 
WHERE id = $1;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds 
ORDER BY last_fetched_at ASC NULLS FIRST LIMIT 1;

-- name: GetFeedByName :one
SELECT * FROM feeds WHERE name = $1;

-- name: GetFeedByURL :one
SELECT * FROM feeds WHERE url = $1;

-- name: GetFeedsByUserName :many
SELECT f.name, f.url 
FROM feeds f 
INNER JOIN users u ON f.user_id = u.id 
WHERE u.name = $1;

-- name: GetAllFeeds :many
SELECT * FROM feeds;

-- name: GetFeedCount :one
SELECT count(*) FROM feeds;

-- name: DeleteAllFeeds :exec
DELETE FROM feeds;
