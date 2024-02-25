-- name: ListMessages :many
SELECT * FROM messages;

-- name: InsertMessage :one
INSERT INTO messages (message) VALUES ($1) RETURNING *;