package main

import (
    "fmt"
    "net/http"
    
    "github.com/a-h/templ"
    "gchalakov.com/TeamForger/pages/signup"
)

func main() {
	http.Handle("/", http.RedirectHandler("/signup", http.StatusSeeOther))
	http.Handle("/signup", templ.Handler(signup.SignUp()))
	http.HandleFunc("/process-signup", func(w http.ResponseWriter, r *http.Request){
		email := r.FormValue("email")
		password := r.FormValue("password")
		repeatedPassword := r.FormValue("repeatedPassword")

		signup.ValidateForm(w, r, email, password, repeatedPassword)
		// Create account
		hashedPassword := signup.HashPassword(email, password)
		conn, err := Connect()
		if err != nil {
			fmt.Println("Could not establish a connection with the DB. Error: %s", err)
			return
		}
		signup.CreateAccount(conn, email, hashedPassword)
	})

	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
