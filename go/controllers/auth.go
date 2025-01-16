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

type loginPage struct {
	Errors   []string
	Login    string
	Register string
}

func loginForm(w http.ResponseWriter, r *http.Request) {
	sess := r.Context().Value(hnContextKey("sess")).(sessions.Session)
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

func register(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	registerToken := r.PostFormValue("register-token")
	registerErrors := make([]string, 0)
	isValid := validateUsername(username)
	ctx := r.Context()
	sess := ctx.Value(hnContextKey("sess")).(sessions.Session)
	token := ctx.Value(hnContextKey("token")).(string)
	pageData := ctx.Value(hnContextKey("page_data")).(*loginPage)
	if !isValid {
		registerErrors = append(registerErrors, "Enter a valid username. This value may contain only letters, numbers, and @/./+/-/_ characters.")
	}
	if password == "" {
		registerErrors = append(registerErrors, "Empty passwords are not allowed")
	}
	if registerToken != token {
		registerErrors = append(registerErrors, "Invalid token")
	}
	if len(registerErrors) > 0 {
		pageData.Errors = registerErrors
		err := renderTemplate(w, "login", pageData)
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
			pageData.Errors = append(pageData.Errors, err.Error())
			renderTemplate(w, "login", pageData)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	sess.Set("user", user)
	sess.Delete("register_token")
	sess.Delete("login_token")
	http.Redirect(w, r, "/", http.StatusFound)
}

func login(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	loginToken := r.PostFormValue("login-token")
	loginErrors := make([]string, 0)
	ctx := r.Context()
	sess := ctx.Value(hnContextKey("sess")).(sessions.Session)
	token := ctx.Value(hnContextKey("token")).(string)
	pageData := ctx.Value(hnContextKey("page_data")).(*loginPage)
	if username == "" || password == "" {
		loginErrors = append(loginErrors, "Empty username or password not allowed")
	}
	if loginToken != token {
		loginErrors = append(loginErrors, "Invalid token")
	}
	if len(loginErrors) > 0 {
		pageData.Errors = loginErrors
		err := renderTemplate(w, "login", pageData)
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
		pageData.Errors = loginErrors
		err := renderTemplate(w, "login", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	sess.Set("user", user)
	sess.Delete("register_token")
	sess.Delete("login_token")
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

func logout(w http.ResponseWriter, r *http.Request) {
	sess := r.Context().Value(hnContextKey("sess")).(sessions.Session)
	user := sess.Get("user")
	if user != nil {
		sessions.GlobalSessions.SessionDestroy(w, r)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
