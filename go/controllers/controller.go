package controllers

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"
	"unicode"

	"github.com/goodyduru/go-news/models"
)

var templates map[string]*template.Template

type hnContextKey string

type form struct {
	Token string
	Goto  string
}

type pageData struct {
	Errors      []string
	CurrentUser *models.User
	User        *models.User
	Posts       []models.Post
	Form        form
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
	templates["reply"] = template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("views/base.html", "views/reply.html"))
	mux := http.NewServeMux()

	// auth
	mux.Handle("GET /register", defaultHandler(http.HandlerFunc(loginForm)))
	mux.Handle("POST /register", authHandler(http.HandlerFunc(register)))
	mux.Handle("GET /login", defaultHandler(http.HandlerFunc(loginForm)))
	mux.Handle("POST /login", authHandler(http.HandlerFunc(login)))
	mux.Handle("GET /logout", defaultHandler(http.HandlerFunc(logout)))

	// post
	mux.Handle("GET /submit", loginRequired(http.HandlerFunc(submitForm)))
	mux.Handle("POST /submit", loginRequired(http.HandlerFunc(submit)))
	mux.Handle("GET /item", checkItemID(http.HandlerFunc(single)))
	mux.Handle("POST /comment", loginRequired(http.HandlerFunc(reply)))
	mux.Handle("GET /reply", loginRequired(http.HandlerFunc(replyForm)))
	mux.Handle("GET /vote", loginRequired(checkItemID(http.HandlerFunc(voteOrFlag))))
	mux.Handle("GET /flag", loginRequired(checkItemID(http.HandlerFunc(voteOrFlag))))

	// home
	mux.Handle("GET /", defaultHandler(http.HandlerFunc(all)))

	// user
	mux.Handle("GET /user", checkUserID(http.HandlerFunc(profile)))
	mux.Handle("POST /user", loginRequired(http.HandlerFunc(updateProfile)))
	mux.Handle("GET /submitted", checkUserID(http.HandlerFunc(submitted)))
	mux.Handle("GET /favorites", checkUserID(http.HandlerFunc(favorites)))
	mux.Handle("GET /threads", checkUserID(http.HandlerFunc(comments)))
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
