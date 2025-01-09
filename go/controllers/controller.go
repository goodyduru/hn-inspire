package controllers

import (
	"html/template"
	"net/http"
)

var templates map[string]*template.Template

func Setup() *http.ServeMux {
	templates = make(map[string]*template.Template)
	templates["login"] = template.Must(template.ParseFiles("views/login.html"))
	templates["index"] = template.Must(template.ParseFiles("views/base.html", "views/index.html"))
	mux := http.NewServeMux()

	// auth
	mux.HandleFunc("GET /register", loginForm)
	mux.HandleFunc("POST /register", register)
	mux.HandleFunc("GET /login", loginForm)
	mux.HandleFunc("POST /login", login)
	return mux
}

func renderTemplate(w http.ResponseWriter, templateName string, data any) error {
	err := templates[templateName].Execute(w, data)
	return err
}
