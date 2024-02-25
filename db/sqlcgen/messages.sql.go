// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: messages.sql

package sqlcgen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const insertMessage = `-- name: InsertMessage :one
INSERT INTO messages (message) VALUES ($1) RETURNING id, message, created_at
`

func (q *Queries) InsertMessage(ctx context.Context, message pgtype.Text) (Message, error) {
	row := q.db.QueryRow(ctx, insertMessage, message)
	var i Message
	err := row.Scan(&i.ID, &i.Message, &i.CreatedAt)
	return i, err
}

const listMessages = `-- name: ListMessages :many
SELECT id, message, created_at FROM messages
`

func (q *Queries) ListMessages(ctx context.Context) ([]Message, error) {
	rows, err := q.db.Query(ctx, listMessages)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Message
	for rows.Next() {
		var i Message
		if err := rows.Scan(&i.ID, &i.Message, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}