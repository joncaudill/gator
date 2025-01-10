-- +goose Up
CREATE TABLE users (
  id uuid PRIMARY KEY,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  name TEXT NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE users;