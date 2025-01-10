package controllers

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/goodyduru/go-news/models"
	"github.com/goodyduru/go-news/sessions"
)

var templates map[string]*template.Template

func Setup() *http.ServeMux {
	templates = make(map[string]*template.Template)
	templates["login"] = template.Must(template.ParseFiles("views/login.html"))
	templates["index"] = template.Must(template.ParseFiles("views/base.html", "views/index.html"))
	mux := http.NewServeMux()

	// auth
	mux.HandleFunc("GET /register", loginForm)
	mux.HandleFunc("POST /register", authHandler(register))
	mux.HandleFunc("GET /login", loginForm)
	mux.HandleFunc("POST /login", authHandler(login))

	// home
	mux.HandleFunc("GET /", home)
	return mux
}

func renderTemplate(w http.ResponseWriter, templateName string, data any) error {
	err := templates[templateName].Execute(w, data)
	return err
}

func home(w http.ResponseWriter, r *http.Request) {
	sess := sessions.StartSession(w, r)
	user := sess.Get("user")
	if user != nil {
		user = user.(*models.User)
	}
	renderTemplate(w, "index", user)
}

func generateToken() string {
	h := md5.New()
	io.WriteString(h, strconv.FormatInt(time.Now().UnixNano(), 10))
	io.WriteString(h, "examplexxxx....")
	token := fmt.Sprintf("%x", h.Sum(nil))
	return token
}
