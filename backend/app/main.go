package main

import (
	"fmt"
	"net/http"
	"log"
	"time"
	"context"
	
	"github.com/a-h/templ"
	"teamforger/backend/pages/signup"
	"teamforger/backend/pages/signin"
	"teamforger/backend/pages/home"
	"teamforger/backend/pages/uploadCV"
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
			log.Printf("Authorization failed: %v", err)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
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
	http.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request){
		// Connect to the DB
		conn, err := core.Connect()
		if err != nil {
			log.Printf("Database connection failed: %v", err)
			http.Redirect(w, r, "/signin?error=databaseError", http.StatusSeeOther)
			return
		}
		// Make sure to close the connection when the function exits
		defer conn.Close(context.Background())

		if err := core.Authorize(conn, r); err == nil {
			log.Println("User already signed in. Redirecting to home.")
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
		templ.Handler(signin.SignIn()).ServeHTTP(w, r)
	})
	
	http.HandleFunc("/process-signin", func(w http.ResponseWriter, r *http.Request){
		var user core.User
		user.Email = r.FormValue("email")
		user.Password = r.FormValue("password")
		user.PasswordHash = core.HashPassword(user.Password)

		// Form validations
		if urlParam, err := core.ValidateEmail(user.Email); err != nil {
			http.Redirect(w, r, "/signin?error=" + urlParam, http.StatusSeeOther)
			return
		}

		// Connect to the DB
		conn, err := core.Connect()
		if err != nil {
			log.Printf("Database connection failed: %v", err)
			http.Redirect(w, r, "/signin?error=databaseError", http.StatusSeeOther)
			return
		}
		// Make sure to close the connection when the function exits
		defer conn.Close(context.Background())

		// Log into an account
		// Get the user by their email
		userDB, err := core.GetUserData(conn, user.Email); 
		if err != nil {
			log.Printf("No user found with email '%s': %v", user.Email, err)
			http.Redirect(w, r, "/signin?error=emailNotFound", http.StatusSeeOther)
			return
		}

		if err := core.CheckPasswordHash(user.Password, userDB.PasswordHash); err != nil {
			log.Printf("Wrong password: %v", err)
			http.Redirect(w, r, "/signin?error=wrongPassword", http.StatusSeeOther)
			return
		}
		// Success
		// Create a session token
		user.SessionToken, err = core.GenerateToken(32)
		if err != nil {
			log.Printf("Failed to generate a SessionToken: %v", err)
			http.Redirect(w, r, "/signin?error=tokenGenerationFailed", http.StatusSeeOther)
			return
		}
		// Create a csrf token
		user.CSRFToken, err = core.GenerateToken(32)
		if err != nil {
			log.Printf("Failed to generate a CSRFToken: %v", err)
			http.Redirect(w, r, "/signin?error=tokenGenerationFailed", http.StatusSeeOther)
			return
		}
		// Set session cookie
		http.SetCookie(w, &http.Cookie {
			Name: "session_token",
			Value: user.SessionToken,
			Expires: time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		})

		// Set csrf token in a cookie
		http.SetCookie(w, &http.Cookie {
			Name: "csrf_token",
			Value: user.CSRFToken,
			Expires: time.Now().Add(24 * time.Hour),
			HttpOnly: false,
		})

		// Set the user email in a cookie
		http.SetCookie(w, &http.Cookie{
			Name: "user_email",
			Value: user.Email,
			Expires: time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		})

		// Update tokens in the DB
		if err := core.UpdateUserTokens(conn, user); err != nil {
			log.Printf("Failed to update tokens in the DB: %v", err)
			http.Redirect(w, r, "/signin?error=tokenGenerationFailed", http.StatusSeeOther)
			return
		}
		// Redirect to home page
		http.Redirect(w, r, "/home?success=welcomeBack", http.StatusSeeOther)
	})
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request){
		// Connect to the DB
		conn, err := core.Connect()
		if err != nil {
			log.Printf("Database connection failed: %v", err)
			http.Redirect(w, r, "/signup?error=databaseError", http.StatusSeeOther)
			return
		}
		// Make sure to close the connection when the function exits
		defer conn.Close(context.Background())

		if err := core.Authorize(conn, r); err == nil {
			log.Println("User already signed in. Redirecting to home.")
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
		templ.Handler(signup.SignUp()).ServeHTTP(w, r)
	})
	http.HandleFunc("/process-signup", func(w http.ResponseWriter, r *http.Request){
		var user core.User
		user.Name = r.FormValue("name")
		user.Email = r.FormValue("email")
		user.Password = r.FormValue("password")
		user.RepeatedPassword = r.FormValue("repeatedPassword")
		user.PasswordHash = core.HashPassword(user.Password)

		// Form validations
		if urlParam, err := core.ValidateEmail(user.Email); err != nil {
			http.Redirect(w, r, "/signup?error=" + urlParam, http.StatusSeeOther)
			return
		}
		if urlParam, err := core.ValidatePassword(user.Password); err != nil {
			http.Redirect(w, r, "/signup?error=" + urlParam, http.StatusSeeOther)
			return
		}
		if urlParam, err := core.CheckPasswordMatch(user.Password, user.RepeatedPassword); err != nil {
			http.Redirect(w, r, "/signup?error=" + urlParam, http.StatusSeeOther)
			return
		}


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
		if err != nil {
			log.Printf("SessionToken generation failed: %v", err)
			http.Redirect(w, r, "/signup?error=tokenGenerationFailed", http.StatusSeeOther)
			return
		}
		// Create a csrf token
		user.CSRFToken, err = core.GenerateToken(32)
		if err != nil {
			log.Printf("CSRFToken generation failed: %v", err)
			http.Redirect(w, r, "/signup?error=tokenGenerationFailed", http.StatusSeeOther)
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
			// Success
			// Set session cookie
			http.SetCookie(w, &http.Cookie {
				Name: "session_token",
				Value: user.SessionToken,
				Expires: time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			})

			// Set csrf token in a cookie
			http.SetCookie(w, &http.Cookie {
				Name: "csrf_token",
				Value: user.CSRFToken,
				Expires: time.Now().Add(24 * time.Hour),
				HttpOnly: false,
			})

			// Set the user email in a cookie
			http.SetCookie(w, &http.Cookie{
				Name: "user_email",
				Value: user.Email,
				Expires: time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			})
			// Redirect to home page
			http.Redirect(w, r, "/home?success=accountCreated", http.StatusSeeOther)
		}
	})
	http.HandleFunc("/signout", func(w http.ResponseWriter, r *http.Request){
		// Connect to the DB
		conn, err := core.Connect()
		if err != nil {
			log.Printf("Database connection failed: %v", err)
			http.Redirect(w, r, "/signin?error=databaseError", http.StatusSeeOther)
			return
		}
		// Make sure to close the connection when the function exits
		defer conn.Close(context.Background())

		if err := core.Authorize(conn, r); err != nil {
			log.Printf("Authorization failed: %v", err)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		// Clear cookies
		http.SetCookie(w, &http.Cookie {
			Name: "session_token",
			Value: "",
			Expires: time.Now().Add(-time.Hour),
			HttpOnly: true,
		})
		http.SetCookie(w, &http.Cookie {
			Name: "csrf_token",
			Value: "",
			Expires: time.Now().Add(-time.Hour),
			HttpOnly: false,
		})
		http.SetCookie(w, &http.Cookie{
			Name: "user_email",
			Value: "",
			Expires: time.Now().Add(-time.Hour),
			HttpOnly: true,
		})

		// Clear tokens from DB
		var emptyUser core.User
		emptyUser.Email = r.FormValue("email")
		emptyUser.SessionToken = ""
		emptyUser.CSRFToken = ""
		if err := core.UpdateUserTokens(conn, emptyUser); err != nil {
			log.Printf("Failed to update tokens in the DB: %v", err)
			http.Redirect(w, r, "/signin?error=tokenGenerationFailed", http.StatusSeeOther)
			return
		}
		
		// Redirect to signin page
		http.Redirect(w, r, "/signin?success=signedOut", http.StatusSeeOther)
	})
	http.HandleFunc("/uploadCV", func(w http.ResponseWriter, r *http.Request){
		// Connect to the DB
		conn, err := core.Connect()
		if err != nil {
			log.Printf("Database connection failed: %v", err)
			http.Redirect(w, r, "/signin?error=databaseError", http.StatusSeeOther)
			return
		}
		// Make sure to close the connection when the function exits
		defer conn.Close(context.Background())

		// Get user email
		emailCookie, err := r.Cookie("user_email")
		if err != nil {
			log.Printf("User's email is not in the cookie: %v", err)
			http.Redirect(w, r, "/error?error=cookieError", http.StatusSeeOther)
		}
		email := emailCookie.Value

		if err := core.Authorize(conn, r); err != nil {
			log.Printf("Authorization failed: %v", err)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		// Get user data
		user, err := core.GetUserData(conn, email)
		if err != nil {
			log.Printf("Retrieving user details failed: %v", err)
			http.Redirect(w, r, "/error?error=databaseError", http.StatusSeeOther)
		}

		templ.Handler(uploadCV.UploadCV(user)).ServeHTTP(w, r)
	})
	http.HandleFunc("/process-uploadCV", func(w http.ResponseWriter, r *http.Request){
		// Connect to the DB
		conn, err := core.Connect()
		if err != nil {
			log.Printf("Database connection failed: %v", err)
			http.Redirect(w, r, "/signin?error=databaseError", http.StatusSeeOther)
			return
		}
		// Make sure to close the connection when the function exits
		defer conn.Close(context.Background())

		// Authorize
		if err := core.Authorize(conn, r); err != nil {
			log.Printf("Authorization failed: %v", err)
			http.Redirect(w, r, "/home?error=authFailed", http.StatusSeeOther)
			return
		}

		// Get the contents of the DOCX file
		fileContents, err := core.ReceiveFile(w, r)
		if err != nil {
			log.Printf("Could not read uploaded file: %v", err)
			http.Redirect(w, r, "/home?error=fileUploadError", http.StatusSeeOther)
			return
		}

		// Convert the DOCX to Markdown
		markdownContent, err := core.DocxToMarkDown(fileContents)
		if err != nil {
			log.Printf("DOCX conversion failed: %v", err)
			http.Redirect(w, r, "/home?error=docxConversionError", http.StatusSeeOther)
			return
		}

		fmt.Println("Converted Markdown Content:")
		fmt.Println(markdownContent)

		// Get user email
		emailCookie, err := r.Cookie("user_email")
		if err != nil {
			log.Printf("User's email is not in the cookie: %v", err)
			http.Redirect(w, r, "/error?error=cookieError", http.StatusSeeOther)
		}
		email := emailCookie.Value

		// Get user data
		user, err := core.GetUserData(conn, email)
		if err != nil {
			log.Printf("Retrieving user details failed: %v", err)
			http.Redirect(w, r, "/error?error=databaseError", http.StatusSeeOther)
		}

		user.CV = markdownContent
		// Store markdownContent in database
		uploadCV.StoreUserCV(conn, user)
		http.Redirect(w, r, "/home?success=CVConverted", http.StatusSeeOther)
	})

	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
