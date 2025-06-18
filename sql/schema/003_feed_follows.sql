
-- +goose Up
CREATE TABLE feed_follows(
	id BIGINT PRIMARY KEY,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	user_id TEXT NOT NULL,
	feed_id BIGINT NOT NULL,
	FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
	CONSTRAINT feed_user UNIQUE (user_id, feed_id)
);

-- +goose down
DROP TABLE feed_follows;
