package signup

import (
	"fmt"
	"unicode"
	"net/mail"
	"golang.org/x/crypto/bcrypt"

	"context"
	"github.com/jackc/pgx/v5"
)

func ValidateForm(email string, password string, repeatedPassword string) string {
	// Check for empty fields
	if email == "" || password == "" || repeatedPassword == "" {
		return "fieldsNotFilled"
	}

	// Validate email structure
	if _, err := mail.ParseAddress(email); err != nil {
		return "badEmail"
	}

	// Check password match first to avoid unnecessary processing
	if password != repeatedPassword {
		return "passwordsDontMatch"
	}

	// Validate password strength
	count := 0
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, r := range password {
		count++
		// Prevent excessively long passwords (DoS protection)
		if count > 256 {
			return "passwordTooLong"
		}

		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if count < 8 {
		return "shortPassword"
	}
	if !hasUpper {
		return "passwordNoUpper"
	}
	if !hasLower {
		return "passwordNoLower"
	}
	if !hasDigit {
		return "passwordNoDigit"
	}
	if !hasSpecial {
		return "passwordNoSpecial"
	}

	return ""
}

func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		fmt.Println(err)
	}
	return string(hashedPassword)
}

func CheckPasswordHash(password string, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type User struct {
	Id int
	Name string
	Email string
	Password string
	RepeatedPassword string
	PasswordHash string
	IsAdmin bool
}

func CountUsers(conn *pgx.Conn) (int, error) {
	rows, err := conn.Query(context.Background(), "SELECT count(id) FROM users")
	if err != nil {
		return -1, err
	}

	var count int
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return -1, err
		}
	}
	return count, nil
}

func CreateUser (conn *pgx.Conn, user User) error {
	userCount, err := CountUsers(conn)
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

	_, err = tx.Exec(context.Background(), "INSERT INTO users (name, email, passwordHash, isAdmin) VALUES ($1, $2, $3, $4)", user.Name, user.Email, user.PasswordHash, user.IsAdmin)

	if err != nil {
	    return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
	    return err
	}

	return nil
}
