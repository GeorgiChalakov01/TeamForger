package main

import (
	"fmt"
	"net/http"
	"net/mail"
	"golang.org/x/crypto/bcrypt"

	"context"
	"log"
	"github.com/jackc/pgx/v5"
)

func ValidateForm(w http.ResponseWriter, r *http.Request, email string, password string, repeatedPassword string){
	// Check if we have the necessary fields filled
	if email == "" || password == "" || repeatedPassword == "" {
		http.Redirect(w, r, "/signup?error=fieldsNotFilled", http.StatusSeeOther)
		return
	}

	// Check the email address structure
	_, emailError := mail.ParseAddress(email)
	if emailError != nil {
		http.Redirect(w, r, "/signup?error=badEmail", http.StatusSeeOther)
		return
	}

	// Check password streanght
	if len(password) < 8 {
		http.Redirect(w, r, "/signup?error=shortPassword", http.StatusSeeOther)
		return
	}

	// Check whether passwords match
	if password != repeatedPassword {
		http.Redirect(w, r, "/signup?error=passwordsDontMatch", http.StatusSeeOther)
		return
	}
}

func HashPassword(email string, password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		fmt.Println("An error occured while hashing the password. Error: %s", err)
	}
	return string(hashedPassword)
}

func Connect() (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:ChangeMe@teamforger-db-1:5432/developerDB")
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func CreateAccount(conn *pgx.Conn, email string, hashedPassword string) {
	rows, err := conn.Query(context.Background(), "SELECT count(id) FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var count int
		err := rows.Scan(&count)
		if err != nil {
			log.Fatal(err)
		}
		if count == 0 {
			fmt.Println("User needs to be admin")
		} else {
			fmt.Println("User Not the first one. ID: %d", count)
		}
	}
}
