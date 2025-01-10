-- +goose Up
CREATE TABLE feeds (
  id uuid PRIMARY KEY,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  name TEXT NOT NULL UNIQUE,
  url TEXT NOT NULL UNIQUE,
  user_ID uuid NOT NULL 
    references users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;