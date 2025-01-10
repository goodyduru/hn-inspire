package controllers

import (
	"context"
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

type hnContextKey string

func Setup() *http.ServeMux {
	templates = make(map[string]*template.Template)
	templates["login"] = template.Must(template.ParseFiles("views/login.html"))
	templates["index"] = template.Must(template.ParseFiles("views/base.html", "views/index.html"))
	mux := http.NewServeMux()

	// auth
	mux.HandleFunc("GET /register", defaultHandler(loginForm))
	mux.HandleFunc("POST /register", authHandler(register))
	mux.HandleFunc("GET /login", defaultHandler(loginForm))
	mux.HandleFunc("POST /login", authHandler(login))

	// home
	mux.HandleFunc("GET /", defaultHandler(home))
	return mux
}

func renderTemplate(w http.ResponseWriter, templateName string, data any) error {
	err := templates[templateName].Execute(w, data)
	return err
}

func home(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	sess := ctx.Value(hnContextKey("sess")).(sessions.Session)
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

func defaultHandler(fn func(context.Context, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		sess := sessions.StartSession(w, r)
		ctx = context.WithValue(ctx, hnContextKey("sess"), sess)
		fn(ctx, w, r)
	}
}
