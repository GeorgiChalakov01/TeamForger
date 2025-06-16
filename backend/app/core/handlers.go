package core

import (
	"context"
	"log"
	"net/http"
	"github.com/jackc/pgx/v5"
)

func WithDBConnection(handler func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := Connect()
		if err != nil {
			log.Printf("Database connection failed: %v", err)
			http.Redirect(w, r, "/signin?error=databaseError", http.StatusSeeOther)
			return
		}
		defer conn.Close(context.Background())
		handler(w, r, conn)
	}
}

func WithAuthorization(handler func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, user User)) http.HandlerFunc {
	return WithDBConnection(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
		if err := Authorize(conn, r); err != nil {
			log.Printf("Authorization failed: %v", err)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		emailCookie, err := r.Cookie("user_email")
		if err != nil {
			log.Printf("User's email is not in the cookie: %v", err)
			http.Redirect(w, r, "/signin?error=cookieError", http.StatusSeeOther)
			return
		}

		user, err := GetUserData(conn, emailCookie.Value)
		if err != nil {
			log.Printf("Retrieving user details failed: %v", err)
			http.Redirect(w, r, "/signin?error=databaseError", http.StatusSeeOther)
			return
		}

		handler(w, r, conn, user)
	})
}

func RedirectIfAuthorized(conn *pgx.Conn, w http.ResponseWriter, r *http.Request, redirectPath string) bool {
	if err := Authorize(conn, r); err == nil {
		log.Println("User already signed in. Redirecting to", redirectPath)
		http.Redirect(w, r, redirectPath, http.StatusSeeOther)
		return true
	}
	return false
}
