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
	"unicode"

	"github.com/goodyduru/go-news/models"
	"github.com/goodyduru/go-news/sessions"
)

var templates map[string]*template.Template

type hnContextKey string
type formToken = string

type pageData struct {
	Errors      []string
	CurrentUser *models.User
	User        *models.User
	Posts       []models.Post
	Form        formToken
	Title       string
}

func pluralize(count int) string {
	if count > 1 {
		return "s"
	}
	return ""
}

func Setup() *http.ServeMux {
	funcMap := template.FuncMap{
		"pluralize": pluralize,
		"timesince": func(saveTime time.Time) string {
			duration := time.Since(saveTime)
			var formattedTime string
			if duration.Seconds() < 60 {
				s := int(duration.Seconds())
				formattedTime = fmt.Sprintf("%d second%s", s, pluralize(s))
			} else if duration.Minutes() < 60 {
				m := int(duration.Minutes())
				formattedTime = fmt.Sprintf("%d minute%s", m, pluralize(m))
			} else if duration.Hours() < 24 {
				h := int(duration.Hours())
				formattedTime = fmt.Sprintf("%d hour%s", h, pluralize(h))
			}
			if formattedTime != "" {
				return formattedTime
			}
			days := int(duration.Hours()) / 24
			if days > 365 {
				formattedTime = fmt.Sprintf("%d year%s", days/365, pluralize(days/365))
			} else if days > 31 {
				formattedTime = fmt.Sprintf("%d month%s", days/31, pluralize(days/31))
			} else if days > 7 {
				formattedTime = fmt.Sprintf("%d week%s", days/7, pluralize(days/7))
			} else {
				formattedTime = fmt.Sprintf("%d day%s", days, pluralize(days))
			}
			return formattedTime
		},
		"indent": func(level int) string {
			if level > 1 {
				return fmt.Sprintf("margin-left: %dem", level-1)
			}
			return ""
		},
		"truncatewords": func(sentence string, max int) string {
			len := 0
			count := 0
			start := 0
			prev := false
			for i, r := range sentence {
				isSpace := unicode.IsSpace(r)
				if isSpace && count > 0 && !prev {
					len++
				} else if !isSpace {
					if count == 0 {
						start = i
					}
					count++
				}
				prev = isSpace
				if len >= max {
					return sentence[start:i] + "..."
				}
			}
			return sentence
		},
	}
	templates = make(map[string]*template.Template)
	templates["login"] = template.Must(template.ParseFiles("views/login.html"))
	templates["index"] = template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("views/base.html", "views/index.html"))
	templates["submit"] = template.Must(template.ParseFiles("views/base.html", "views/submit.html"))
	templates["user"] = template.Must(template.ParseFiles("views/base.html", "views/profile.html"))
	templates["mixed"] = template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("views/base.html", "views/mixed.html"))
	templates["threads"] = template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("views/base.html", "views/comments.html"))
	templates["single"] = template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("views/base.html", "views/single.html"))
	mux := http.NewServeMux()

	// auth
	mux.HandleFunc("GET /register", defaultHandler(loginForm))
	mux.HandleFunc("POST /register", authHandler(register))
	mux.HandleFunc("GET /login", defaultHandler(loginForm))
	mux.HandleFunc("POST /login", authHandler(login))
	mux.HandleFunc("GET /logout", defaultHandler(logout))

	// post
	mux.HandleFunc("GET /submit", loginRequired(submitForm))
	mux.HandleFunc("POST /submit", loginRequired(submit))
	mux.HandleFunc("GET /item", checkItemID(single))

	// home
	mux.HandleFunc("GET /", defaultHandler(all))

	// user
	mux.HandleFunc("GET /user", checkUserID(profile))
	mux.HandleFunc("POST /user", loginRequired(updateProfile))
	mux.HandleFunc("GET /submitted", checkUserID(submitted))
	mux.HandleFunc("GET /favorites", checkUserID(favorites))
	mux.HandleFunc("GET /threads", checkUserID(comments))
	return mux
}

func renderTemplate(w http.ResponseWriter, pageName string, data any) error {
	err := templates[pageName].Execute(w, data)
	return err
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
		ctx := r.Context()
		sess := sessions.GlobalSessions.SessionStart(w, r)
		ctx = context.WithValue(ctx, hnContextKey("sess"), sess)
		fn(ctx, w, r)
	}
}

func loginRequired(fn func(context.Context, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := sessions.GlobalSessions.SessionStart(w, r)
		user := sess.Get("user")
		if user == nil {
			http.Error(w, "You are not logged in", http.StatusUnauthorized)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, hnContextKey("sess"), sess)
		fn(ctx, w, r)
	}
}

func checkUserID(fn func(context.Context, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("id")
		if user == "" {
			http.Error(w, "No such user exist", http.StatusNotFound)
			return
		}
		u := models.User{Username: user}
		if err := u.ReadByUsername(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if u.ID == 0 {
			http.Error(w, "No such user exist", http.StatusNotFound)
			return
		}
		ctx := r.Context()
		sess := sessions.GlobalSessions.SessionStart(w, r)
		ctx = context.WithValue(ctx, hnContextKey("sess"), sess)
		ctx = context.WithValue(ctx, hnContextKey("user"), &u)
		fn(ctx, w, r)
	}
}

func checkItemID(fn func(context.Context, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		post := r.URL.Query().Get("id")
		if post == "" {
			http.Error(w, "No such post exist", http.StatusNotFound)
			return
		}
		id, err := strconv.ParseInt(post, 10, 64)
		if err != nil {
			http.Error(w, "No such post exist", http.StatusNotFound)
		}
		sess := sessions.GlobalSessions.SessionStart(w, r)
		p := models.Post{ID: int(id)}
		user := sess.Get("user")
		if user != nil {
			u := user.(*models.User)
			err = p.GetAuth(u.ID)
		} else {
			err = p.Get()
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if p.AuthorID == 0 {
			http.Error(w, "No such post exist", http.StatusNotFound)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, hnContextKey("sess"), sess)
		ctx = context.WithValue(ctx, hnContextKey("post"), p)
		fn(ctx, w, r)
	}
}
