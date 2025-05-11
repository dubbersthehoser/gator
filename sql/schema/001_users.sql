-- +goose Up
CREATE TABLE users (
	id TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	name TEXT NOT NULL
);

-- +goose Down
DROP TABLE users;
