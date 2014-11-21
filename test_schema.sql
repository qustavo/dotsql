-- name: create-users-table
CREATE TABLE users (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	name VARCHAR(255),
	email VARCHAR(255)
);

-- name: create-user
INSERT INTO users (name, email) VALUES(?, ?)

-- name: find-one-user-by-email
SELECT id,name,email FROM users WHERE email = ? LIMIT 1

-- name: drop-users-table
DROP TABLE users
