// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package sqlcgen

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Message struct {
	ID        pgtype.UUID
	Message   pgtype.Text
	CreatedAt pgtype.Timestamptz
}

type SchemaMigration struct {
	Version string
}
