package main

import (
	"fmt"
	"net/http"
	"time"
	
	"github.com/a-h/templ"
	"github.com/jackc/pgx/v5"
	"teamforger/backend/core"
	"teamforger/backend/pages/signup"
	"teamforger/backend/pages/signin"
	"teamforger/backend/pages/home"
	"teamforger/backend/pages/uploadCV"
)

func main() {
	http.Handle("/", http.RedirectHandler("/signup", http.StatusSeeOther))
	
	http.HandleFunc("/home", core.WithAuthorization(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, user core.User) {
		templ.Handler(home.Home(user)).ServeHTTP(w, r)
	}))
	
	http.HandleFunc("/signin", core.WithDBConnection(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
		if core.RedirectIfAuthorized(conn, w, r, "/home") {
			return
		}
		templ.Handler(signin.SignIn()).ServeHTTP(w, r)
	}))
	
	http.HandleFunc("/process-signin", core.WithDBConnection(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
		user := core.User{
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
		}
		user.PasswordHash = core.HashPassword(user.Password)

		if urlParam, err := core.ValidateEmail(user.Email); err != nil {
			http.Redirect(w, r, "/signin?error="+urlParam, http.StatusSeeOther)
			return
		}

		userDB, err := core.GetUserData(conn, user.Email)
		if err != nil {
			http.Redirect(w, r, "/signin?error=emailNotFound", http.StatusSeeOther)
			return
		}

		if err := core.CheckPasswordHash(user.Password, userDB.PasswordHash); err != nil {
			http.Redirect(w, r, "/signin?error=wrongPassword", http.StatusSeeOther)
			return
		}

		if err := core.GenerateAndSetTokens(w, &user); err != nil {
			http.Redirect(w, r, "/signin?error=tokenGenerationFailed", http.StatusSeeOther)
			return
		}

		if err := core.UpdateUserTokens(conn, user); err != nil {
			http.Redirect(w, r, "/signin?error=tokenUpdateFailed", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/home?success=welcomeBack", http.StatusSeeOther)
	}))
	
	http.HandleFunc("/signup", core.WithDBConnection(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
		if core.RedirectIfAuthorized(conn, w, r, "/home") {
			return
		}
		templ.Handler(signup.SignUp()).ServeHTTP(w, r)
	}))
	
	http.HandleFunc("/process-signup", core.WithDBConnection(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
		user := core.User{
			Name:             r.FormValue("name"),
			Email:            r.FormValue("email"),
			Password:         r.FormValue("password"),
			RepeatedPassword: r.FormValue("repeatedPassword"),
		}
		user.PasswordHash = core.HashPassword(user.Password)

		if urlParam, err := core.ValidateEmail(user.Email); err != nil {
			http.Redirect(w, r, "/signup?error="+urlParam, http.StatusSeeOther)
			return
		}
		if urlParam, err := core.ValidatePassword(user.Password); err != nil {
			http.Redirect(w, r, "/signup?error="+urlParam, http.StatusSeeOther)
			return
		}
		if urlParam, err := core.CheckPasswordMatch(user.Password, user.RepeatedPassword); err != nil {
			http.Redirect(w, r, "/signup?error="+urlParam, http.StatusSeeOther)
			return
		}

		if err := core.GenerateAndSetTokens(w, &user); err != nil {
			http.Redirect(w, r, "/signup?error=tokenGenerationFailed", http.StatusSeeOther)
			return
		}

		if err := signup.CreateUser(conn, user); err != nil {
			if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)" {
				http.Redirect(w, r, "/signup?error=duplicateEmail", http.StatusSeeOther)
			} else {
				http.Redirect(w, r, "/signup?error=createAccountError", http.StatusSeeOther)
			}
			return
		}

		http.Redirect(w, r, "/home?success=accountCreated", http.StatusSeeOther)
	}))
	
	http.HandleFunc("/signout", core.WithAuthorization(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, user core.User) {
		// Clear cookies
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   "",
			Expires: time.Now().Add(-time.Hour),
		})
		http.SetCookie(w, &http.Cookie{
			Name:    "csrf_token",
			Value:   "",
			Expires: time.Now().Add(-time.Hour),
		})
		http.SetCookie(w, &http.Cookie{
			Name:    "user_email",
			Value:   "",
			Expires: time.Now().Add(-time.Hour),
		})

		// Clear tokens from DB
		emptyUser := core.User{
			Email:        user.Email,
			SessionToken: "",
			CSRFToken:    "",
		}
		if err := core.UpdateUserTokens(conn, emptyUser); err != nil {
			http.Redirect(w, r, "/signin?error=tokenClearFailed", http.StatusSeeOther)
			return
		}
		
		http.Redirect(w, r, "/signin?success=signedOut", http.StatusSeeOther)
	}))
	
	http.HandleFunc("/uploadCV", core.WithAuthorization(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, user core.User) {
		templ.Handler(uploadCV.UploadCV(user)).ServeHTTP(w, r)
	}))
	
	http.HandleFunc("/process-uploadCV", core.WithAuthorization(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, user core.User) {
		fileContents, err := core.ReceiveFile(w, r)
		if err != nil {
			http.Redirect(w, r, "/home?error=fileUploadError", http.StatusSeeOther)
			return
		}

		markdownContent, err := core.DocxToMarkDown(fileContents)
		if err != nil {
			http.Redirect(w, r, "/home?error=docxConversionError", http.StatusSeeOther)
			return
		}

		user.CV = markdownContent
		if err := uploadCV.StoreUserCV(conn, user); err != nil {
			http.Redirect(w, r, "/home?error=cvStorageFailed", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/home?success=CVConverted", http.StatusSeeOther)
	}))


	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
