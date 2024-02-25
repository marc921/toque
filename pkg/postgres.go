package pkg

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Connect establishes a connection with the database and returns it.
func NewPostgresConnection(ctx context.Context, connString string) (*pgx.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("pgx.Connect: %w", err)
	}

	// Ping the database.
	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("conn.Ping: %w", err)
	}
	return conn, nil
}
