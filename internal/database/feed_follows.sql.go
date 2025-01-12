// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: feed_follows.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createFeedFollow = `-- name: CreateFeedFollow :one
INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING id, created_at, updated_at, user_id, feed_id
`

type CreateFeedFollowParams struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    uuid.UUID
	FeedID    uuid.UUID
}

// WITH inserted_feed_follow AS (
func (q *Queries) CreateFeedFollow(ctx context.Context, arg CreateFeedFollowParams) (FeedFollow, error) {
	row := q.db.QueryRowContext(ctx, createFeedFollow,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.UserID,
		arg.FeedID,
	)
	var i FeedFollow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
		&i.FeedID,
	)
	return i, err
}

const getFeedFollowsForUser = `-- name: GetFeedFollowsForUser :many

SELECT ff.id, ff.created_at, feeds.name AS feed_name, users.name AS user_name
FROM feed_follows ff
INNER JOIN feeds ON ff.feed_id = feeds.id
INNER JOIN users ON ff.user_id = users.id
WHERE ff.user_id = $1
`

type GetFeedFollowsForUserRow struct {
	ID        uuid.UUID
	CreatedAt time.Time
	FeedName  string
	UserName  string
}

// )
// SELECT
//
//	inserted_feed_follow.*,
//	feeds.name as feed_name,
//	users.name as user_name
//
// FROM inserted_feed_follow
// INNER JOIN feeds ON inserted_feed_follow.feed_id = feeds.id
// INNER JOIN users ON inserted_feed_follow.user_id = users.id;
func (q *Queries) GetFeedFollowsForUser(ctx context.Context, userID uuid.UUID) ([]GetFeedFollowsForUserRow, error) {
	rows, err := q.db.QueryContext(ctx, getFeedFollowsForUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetFeedFollowsForUserRow
	for rows.Next() {
		var i GetFeedFollowsForUserRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.FeedName,
			&i.UserName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const resetFeedFollows = `-- name: ResetFeedFollows :exec
DELETE FROM feed_follows
`

func (q *Queries) ResetFeedFollows(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, resetFeedFollows)
	return err
}