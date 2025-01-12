-- +goose Up
CREATE TABLE posts (
  id uuid PRIMARY KEY,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  title TEXT NOT NULL DEFAULT '',
  url TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  published_at TIMESTAMP NOT NULL DEFAULT NOW(),
  feed_id uuid NOT NULL 
    references feeds(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;