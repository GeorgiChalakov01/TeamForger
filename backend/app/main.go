package main

import (
	"fmt"
	"net/http"
	"log"
	"time"
	"context"
	
	"github.com/a-h/templ"
	"teamforger/backend/pages/signup"
	"teamforger/backend/pages/home"
	"teamforger/backend/core"
)

func main() {
	http.Handle("/", http.RedirectHandler("/signup", http.StatusSeeOther))
	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request){
		// Connect to the DB
		conn, err := core.Connect()
		// Make sure to close the connection when the function exits
		defer conn.Close(context.Background())

		if err != nil {
			log.Printf("Database connection failed: %v", err)
			http.Redirect(w, r, "/error?error=databaseError", http.StatusSeeOther)
			return
		}
		if err := core.Authorize(conn, r); err != nil {
			er := http.StatusUnauthorized
			http.Error(w, "Unauthorized", er)
			return
		}

		emailCookie, err := r.Cookie("user_email")
		if err != nil {
			log.Printf("User's email is not in the cookie: %v", err)
			http.Redirect(w, r, "/error?error=cookieError", http.StatusSeeOther)
		}
		email := emailCookie.Value

		user, err := core.GetUserData(conn, email)
		if err != nil {
			log.Printf("Retrieving user details failed: %v", err)
			http.Redirect(w, r, "/error?error=databaseError", http.StatusSeeOther)
		}

		templ.Handler(home.Home(user)).ServeHTTP(w, r)
	})
	http.Handle("/signup", templ.Handler(signup.SignUp()))
	http.HandleFunc("/process-signup", func(w http.ResponseWriter, r *http.Request){
		var user core.User
		user.Email = r.FormValue("email")
		user.Password = r.FormValue("password")
		user.RepeatedPassword = r.FormValue("repeatedPassword")

		msg := signup.ValidateForm(user.Email, user.Password, user.RepeatedPassword)
		if msg != "" {
			http.Redirect(w, r, "/signup?error=" + msg, http.StatusSeeOther)
			return
		} else {
			// Hash the Password
			user.PasswordHash = signup.HashPassword(user.Password)

			// Connect to the DB
			conn, err := core.Connect()
			// Make sure to close the connection when the function exits
			defer conn.Close(context.Background())

			if err != nil {
				log.Printf("Database connection failed: %v", err)
				http.Redirect(w, r, "/signup?error=databaseError", http.StatusSeeOther)
				return
			}
			// Create a session token
			user.SessionToken, err = core.GenerateToken(32)
			// Create a csrf token
			user.CSRFToken, err = core.GenerateToken(32)
			if err != nil {
				http.Redirect(w, r, "/login?error=tokenGenerationFailed", http.StatusSeeOther)
			}
			// Create account
			if err := signup.CreateUser(conn, user); err != nil {
				if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)" {
					log.Printf("This email is already used: %v", err)
					http.Redirect(w, r, "/signup?error=duplicateEmail", http.StatusSeeOther)
				} else {
					log.Printf("An error occured while creating the user account: %v", err)
					http.Redirect(w, r, "/signup?error=createAccountError", http.StatusSeeOther)
				}
				return
			} else {
				// Success
				// Set session cookie
				http.SetCookie(w, &http.Cookie {
					Name: "session_token",
					Value: user.SessionToken,
					Expires: time.Now().Add(23 * time.Hour),
					HttpOnly: true,
				})

				// Set csrf token in a cookie
				http.SetCookie(w, &http.Cookie {
					Name: "csrf_token",
					Value: user.CSRFToken,
					Expires: time.Now().Add(23 * time.Hour),
					HttpOnly: false,
				})

				// Set the user email in a cookie
				http.SetCookie(w, &http.Cookie{
					Name: "user_email",
					Value: user.Email,
					Expires: time.Now().Add(23 * time.Hour),
					HttpOnly: true,
				})
				// Redirect to login page
				http.Redirect(w, r, "/home?success=accountCreated", http.StatusSeeOther)
			}
		}
	})

	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
