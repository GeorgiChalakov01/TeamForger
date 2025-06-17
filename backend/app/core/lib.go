package core

import (
	"fmt"

	"context"
	"errors"
	"net/http"
	"github.com/jackc/pgx/v5"

	"crypto/rand"
	"encoding/base64"
	"log"

	"golang.org/x/crypto/bcrypt"

	// Read file from post request
	"strings"
	"io"
	"os"

	// Docx to MarkDown
	"github.com/zakahan/docx2md"
	"path/filepath"

	"time"
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
	CV string
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

func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		fmt.Println(err)
	}
	return string(hashedPassword)
}

func CheckPasswordHash(password string, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err
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
		"SELECT id, name, email, passwordHash, sessionToken, csrfToken, isAdmin, cv FROM users WHERE email=$1", email).Scan(
			&user.Id, &user.Name, &user.Email, &user.PasswordHash, &user.SessionToken, &user.CSRFToken, &user.IsAdmin, &user.CV)
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
		// CSRFToken := r.Header.Get("X-CSRF-Token") // This was the original but couldn't get it to work
		// Get CSRF token from form value
		CSRFToken := r.FormValue("csrf_token") // Replaced it with this, hope it is good enough.
		if CSRFToken == "" || CSRFToken != user.CSRFToken {
			return AuthError
		}
	}

	return nil
}

func UpdateUserTokens(conn *pgx.Conn, user User) error {
	// Start a transaction
	tx, err := conn.Begin(context.Background())
	if err != nil {
		return err
	}
	// Rollback is safe to call even if the tx is already closed, so if
	// the tx commits successfully, this is a no-op
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), "UPDATE users SET sessionToken = $1, csrfToken = $2 WHERE email = $3", user.SessionToken, user.CSRFToken, user.Email)

	if err != nil {
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func ReceiveFile(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	// Check content type first
	if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "multipart/form-data") {
		return nil, fmt.Errorf("expected multipart/form-data, got %s", ct)
	}

	// Parse form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32 MB max memory
		return nil, fmt.Errorf("failed to parse form: %w", err)
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("file upload error: %w", err)
	}
	defer file.Close()

	// Validate file type - allow DOCX
	contentType := header.Header.Get("Content-Type")
	allowedTypes := []string{
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document", // DOCX
	}

	validType := false
	for _, t := range allowedTypes {
		if contentType == t {
			validType = true
			break
		}
	}

	// Also check file extension as fallback
	ext := filepath.Ext(header.Filename)
	if !validType && ext != ".docx" {
		return nil, errors.New("only DOCX files are allowed")
	}

	// Read file directly into byte slice
	return io.ReadAll(file)
}

func DocxToMarkDown(docxBytes []byte) (string, error) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "docx-convert-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create temp input file
	inputPath := filepath.Join(tmpDir, "input.docx")
	if err := os.WriteFile(inputPath, docxBytes, 0644); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	// Create temp output directory
	outputDir := filepath.Join(tmpDir, "output")
	if err := os.Mkdir(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Convert DOCX to Markdown
	outputPath, mdString, err := docx2md.DocxConvert(inputPath, outputDir)
	if err != nil {
		return "", fmt.Errorf("conversion failed: %w", err)
	}
	
	// Read the full output file (mdString might be truncated for large files)
	if mdString == "" {
		mdBytes, err := os.ReadFile(outputPath)
		if err != nil {
			return "", fmt.Errorf("failed to read output file: %w", err)
		}
		mdString = string(mdBytes)
	}

	return mdString, nil
}

func GenerateAndSetTokens(w http.ResponseWriter, user *User) error {
	var err error
	user.SessionToken, err = GenerateToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate session token: %w", err)
	}

	user.CSRFToken, err = GenerateToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    user.SessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    user.CSRFToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: false,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "user_email",
		Value:    user.Email,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	return nil
}
