package core

import (
	//"fmt"
	"os"

	"context"
	"errors"
	"net/http"
	"github.com/jackc/pgx/v5"

	"crypto/rand"
	"encoding/base64"
	"log"
)

type User struct {
	Id int
	Name string
	Email string
	Password string
	RepeatedPassword string  
	PasswordHash string
	SessionToken string
	CSRFToken string
	IsAdmin bool
}

func Connect() (*pgx.Conn, error) {
	containerName := os.Getenv("DB_CONTAINER_NAME")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PWD")
	schema := os.Getenv("DB_SCHEMA")
	port := os.Getenv("DB_PORT")

	url := "postgres://" + user + ":" + pass + "@" + containerName + ":" + port + "/" + schema

	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("Failed to generate token: %v", err)
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func GetUserData(conn *pgx.Conn, email string) (User, error) {
	var user User
	err := conn.QueryRow(
		context.Background(),
		"SELECT id, name, email, passwordHash, sessionToken, csrfToken, isAdmin FROM users WHERE email=$1", email).Scan(
			&user.Id, &user.Name, &user.Email, &user.PasswordHash, &user.SessionToken, &user.CSRFToken, &user.IsAdmin)
	if err != nil {
		return user, err
	}
	return user, nil
}

func Authorize(con *pgx.Conn, r *http.Request) error {
	var AuthError = errors.New("Unauthorized")
	emailCookie, err := r.Cookie("user_email")
	if err != nil {
		return AuthError
	}
	email := emailCookie.Value

	user, err := GetUserData(con, email)
	if err != nil {
		return AuthError
	}

	sessionToken, err := r.Cookie("session_token")
	if err != nil || sessionToken.Value == "" || sessionToken.Value != user.SessionToken {
		return AuthError
	}

	// Only require CSRF for non-GET requests
	if r.Method != "GET" {
		CSRFToken := r.Header.Get("X-CSRF-Token")
		if CSRFToken != user.CSRFToken || CSRFToken == "" {
			return AuthError
		}
	}

	return nil
}
