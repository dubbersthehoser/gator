
-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
	INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id) 
	VALUES (
		$1,
		$2,
		$3,
		$4,
		$5
	)
    RETURNING *
) SELECT
	inserted_feed_follow.*,
	f.name AS feed_name,
	u.name AS user_name
FROM inserted_feed_follow
INNER JOIN users u ON inserted_feed_follow.user_id = u.id
INNER JOIN feeds f ON inserted_feed_follow.feed_id = f.id;

-- name: GetFeedFollowsCount :one
SELECT count(*) FROM feed_follows;

-- name: GetFeedFollowsForUser :many
SELECT feed_follows.user_id, feed_follows.feed_id, f.name AS feed_name, u.name AS user_name
FROM feed_follows 
INNER JOIN users u ON feed_follows.user_id = u.id
INNER JOIN feeds f ON feed_follows.feed_id = f.id
WHERE u.name = $1;

-- name: FeedUnfollow :exec
DELETE FROM feed_follows
USING users u, feeds f
WHERE feed_follows.user_id = u.id 
AND feed_follows.feed_id = f.id 
AND u.name = $1 
AND f.url = $2;

