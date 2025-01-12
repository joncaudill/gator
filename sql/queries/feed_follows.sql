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
)
SELECT
    inserted_feed_follow.*,
    feeds.name as feed_name,
    users.name as user_name
FROM inserted_feed_follow
INNER JOIN feeds ON inserted_feed_follow.feed_id = feeds.id
INNER JOIN users ON inserted_feed_follow.user_id = users.id;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows WHERE user_id = $1 AND feed_id = $2;

-- name: GetFeedFollowsForUser :many
SELECT ff.id, ff.created_at, feeds.name AS feed_name, users.name AS user_name
FROM feed_follows ff
INNER JOIN feeds ON ff.feed_id = feeds.id
INNER JOIN users ON ff.user_id = users.id
WHERE ff.user_id = $1;

-- name: ResetFeedFollows :exec
DELETE FROM feed_follows;