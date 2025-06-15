package uploadCV

import (
	"teamforger/backend/core"
	"github.com/jackc/pgx/v5"
	"context"
)

func StoreUserCV(conn *pgx.Conn, user core.User) error {
	// Start a transaction
	tx, err := conn.Begin(context.Background())
	if err != nil {
		return err
	}
	// Rollback is safe to call even if the tx is already closed, so if
	// the tx commits successfully, this is a no-op
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), "UPDATE users SET cv = $1 WHERE email = $2", user.CV, user.Email)

	if err != nil {
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}

