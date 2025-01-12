-- +goose Up
CREATE TABLE feed_follows (
  id uuid PRIMARY KEY,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  user_id uuid NOT NULL 
    references users(id) ON DELETE CASCADE,
  feed_id uuid NOT NULL 
    references feeds(id) ON DELETE CASCADE,
  UNIQUE (user_id, feed_id)
);

-- +goose Down
DROP TABLE feed_follows;