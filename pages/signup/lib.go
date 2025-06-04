package signup

import (
	"net/http"
	"net/mail"
	"crypto/sha256"
)

func ValidateForm(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w, "Invalid Method", er)
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")
	repeatedPassword := r.FormValue("repeatedPassword")

	// Check if we have the necessary fields filled
	if email == "" || password == "" || repeatedPassword == "" {
		http.Redirect(w, r, "/signup?error=fieldsNotFilled", http.StatusSeeOther)
		return
	}

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

func CreateUserAccount(email string, password string) {
	// Hash Password
}
