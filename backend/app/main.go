package main

import (
	"fmt"
	"net/http"
	"log"
	"context"
	
	"github.com/a-h/templ"
	"gchalakov.com/TeamForger/pages/signup"
)

func main() {
	http.Handle("/", http.RedirectHandler("/signup", http.StatusSeeOther))
	http.Handle("/signup", templ.Handler(signup.SignUp()))
	http.HandleFunc("/process-signup", func(w http.ResponseWriter, r *http.Request){
		var user signup.User
		user.Email = r.FormValue("email")
		user.Password = r.FormValue("password")
		user.RepeatedPassword = r.FormValue("repeatedPassword")

		msg := signup.ValidateForm(user.Email, user.Password, user.RepeatedPassword)
		if msg != "" {
			http.Redirect(w, r, "/signup?error=" + msg, http.StatusSeeOther)
			return
		} else {
			// Hash the Password
			user.PasswordHash = signup.HashPassword(user.Email, user.Password)

			// Connect to the DB
			conn, err := Connect()
			// Make sure to close the connection where the function exits
			defer conn.Close(context.Background())

			if err != nil {
				log.Printf("Database connection failed: %v", err)
				http.Redirect(w, r, "/signup?error=databaseError", http.StatusSeeOther)
				return
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
				// Success! Redirect to login page
				http.Redirect(w, r, "/login?success=accountCreated", http.StatusSeeOther)
			}
		}
	})

	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
