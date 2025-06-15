package core

import (
	"errors"
	"unicode"
	"net/mail"
)

func ValidateEmail(email string) (string, error) {
	if email == "" {
		return "emailEmpty", errors.New("email is required")
	}
	
	if _, err := mail.ParseAddress(email); err != nil {
		return "badEmail", errors.New("invalid email format")
	}
	return "", nil
}

func ValidatePassword(password string) (string, error) {
	if password == "" {
		return "passwordEmpty", errors.New("password is required")
	}

	count := 0
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, r := range password {
		count++
		if count > 256 {
			return "passwordTooLong", errors.New("password exceeds maximum length")
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
		return "shortPassword", errors.New("password is too short")
	}
	if !hasUpper {
		return "passwordNoUpper", errors.New("password requires uppercase letter")
	}
	if !hasLower {
		return "passwordNoLower", errors.New("password requires lowercase letter")
	}
	if !hasDigit {
		return "passwordNoDigit", errors.New("password requires digit")
	}
	if !hasSpecial {
		return "passwordNoSpecial", errors.New("password requires special character")
	}

	return "", nil
}

func CheckPasswordMatch(password, repeatedPassword string) (string, error) {
	if password != repeatedPassword {
		return "passwordsDontMatch", errors.New("passwords do not match")
	}
	return "", nil
}
