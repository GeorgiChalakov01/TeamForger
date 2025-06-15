package signup

import (
	"fmt"
	"context"
	"github.com/jackc/pgx/v5"
	"teamforger/backend/core"
)

func CreateUser (conn *pgx.Conn, user core.User) error {
	userCount, err := core.CountUsers(conn)
	if err != nil {
		fmt.Println("Could not count the users. Error: ")
		return err
	}
	if userCount == 0 {
		fmt.Println("User will be created as an admin")
		user.IsAdmin = true
	} else {
		user.IsAdmin = false
	}

	// Start a transaction
	tx, err := conn.Begin(context.Background())
	if err != nil {
	    return err
	}
	// Rollback is safe to call even if the tx is already closed, so if
	// the tx commits successfully, this is a no-op
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), "INSERT INTO users (name, email, passwordHash, sessionToken, csrfToken, isAdmin, cv) VALUES ($1, $2, $3, $4, $5, $6, $7)", user.Name, user.Email, user.PasswordHash, user.SessionToken, user.CSRFToken, user.IsAdmin, "")

	if err != nil {
	    return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
	    return err
	}

	return nil
}
