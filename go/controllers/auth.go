package controllers

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"

	"github.com/goodyduru/go-news/models"
	"github.com/goodyduru/go-news/sessions"
	"golang.org/x/crypto/bcrypt"
)

func loginForm(w http.ResponseWriter, r *http.Request) {
	if err := renderTemplate(w, "login", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	loginErrors := make([]string, 0)
	isValid := validateUsername(username)
	if !isValid {
		loginErrors = append(loginErrors, "Enter a valid username. This value may contain only letters, numbers, and @/./+/-/_ characters.")
	}
	if password == "" {
		loginErrors = append(loginErrors, "Empty passwords are not allowed")
	}
	if len(loginErrors) > 0 {
		err := renderTemplate(w, "login", loginErrors)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	user := &models.User{
		Username: username,
		Password: password,
	}
	if err := user.Create(); err != nil {
		if errors.Is(err, models.ErrNotUnique) {
			renderTemplate(w, "login", err.Error())
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sess := sessions.StartSession(w, r)
	sess.Set("user", user)
	http.Redirect(w, r, "/", http.StatusFound)
}

func login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	loginErrors := make([]string, 0)
	if username == "" || password == "" {
		loginErrors = append(loginErrors, "Empty username or password not allowed")
	}
	if len(loginErrors) > 0 {
		err := renderTemplate(w, "login", loginErrors)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	errMessage := "Please enter a correct username and password."
	user := &models.User{Username: username}
	if err := user.ReadByUsername(); err != nil {
		if err == sql.ErrNoRows {
			loginErrors = append(loginErrors, errMessage)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	isCorrectPassword := verify(password, user.Password)
	if !isCorrectPassword {
		if len(loginErrors) == 0 {
			loginErrors = append(loginErrors, errMessage)
		}
		err := renderTemplate(w, "login", loginErrors)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	sess := sessions.StartSession(w, r)
	sess.Set("user", user)
	http.Redirect(w, r, "/", http.StatusFound)
}

// From Django username validator
func validateUsername(username string) bool {
	m, _ := regexp.MatchString(`^[\w.@+-]+$`, username)
	return m
}

func verify(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
