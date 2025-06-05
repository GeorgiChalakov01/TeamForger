package signup

import (
	"fmt"
	"net/http"
	"net/mail"
	"golang.org/x/crypto/bcrypt"
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
		fmt.Printf("An error occured while hashing the password. Error: %s", err)
	}
	return string(hashedPassword)
}

func CreateAccount(email string, hashedPassword string) {
}
