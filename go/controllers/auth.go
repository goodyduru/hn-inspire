package controllers

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/goodyduru/go-news/models"
	"github.com/goodyduru/go-news/sessions"
	"golang.org/x/crypto/bcrypt"
)

type authData struct {
	pageData loginPage
	token    string
	sess     sessions.Session
}

type loginPage struct {
	Errors   []string
	Login    string
	Register string
}

func loginForm(w http.ResponseWriter, r *http.Request) {
	sess := sessions.StartSession(w, r)
	l := loginPage{
		Login:    generateToken(),
		Register: generateToken(),
	}
	sess.Set("login_token", l.Login)
	sess.Set("register_token", l.Register)
	if err := renderTemplate(w, "login", l); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func authHandler(fn func(http.ResponseWriter, *http.Request, *authData)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		authData := &authData{sess: sessions.StartSession(w, r)}
		registerToken := authData.sess.Get("register_token")
		loginToken := authData.sess.Get("login_token")
		if registerToken == nil || loginToken == nil {
			http.Error(w, "Invalid submission", http.StatusForbidden)
			return
		}
		if strings.Contains(r.URL.Path, "login") {
			authData.token = loginToken.(string)
		} else {
			authData.token = registerToken.(string)
		}
		authData.pageData = loginPage{
			Login:    generateToken(),
			Register: generateToken(),
		}
		authData.sess.Set("login_token", authData.pageData.Login)
		authData.sess.Set("register_token", authData.pageData.Register)
		fn(w, r, authData)
	}
}

func register(w http.ResponseWriter, r *http.Request, a *authData) {
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	registerToken := r.PostFormValue("register-token")
	registerErrors := make([]string, 0)
	isValid := validateUsername(username)
	if !isValid {
		registerErrors = append(registerErrors, "Enter a valid username. This value may contain only letters, numbers, and @/./+/-/_ characters.")
	}
	if password == "" {
		registerErrors = append(registerErrors, "Empty passwords are not allowed")
	}
	if registerToken != a.token {
		registerErrors = append(registerErrors, "Invalid token")
	}
	if len(registerErrors) > 0 {
		a.pageData.Errors = registerErrors
		err := renderTemplate(w, "login", a.pageData)
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
			a.pageData.Errors = append(a.pageData.Errors, err.Error())
			renderTemplate(w, "login", a.pageData)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	a.sess.Set("user", user)
	a.sess.Delete("register_token")
	a.sess.Delete("login_token")
	http.Redirect(w, r, "/", http.StatusFound)
}

func login(w http.ResponseWriter, r *http.Request, a *authData) {
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	loginToken := r.PostFormValue("login-token")
	loginErrors := make([]string, 0)
	if username == "" || password == "" {
		loginErrors = append(loginErrors, "Empty username or password not allowed")
	}
	if loginToken != a.token {
		loginErrors = append(loginErrors, "Invalid token")
	}
	if len(loginErrors) > 0 {
		a.pageData.Errors = loginErrors
		err := renderTemplate(w, "login", a.pageData)
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
		a.pageData.Errors = loginErrors
		err := renderTemplate(w, "login", a.pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	sess := sessions.StartSession(w, r)
	sess.Set("user", user)
	a.sess.Delete("register_token")
	a.sess.Delete("login_token")
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
