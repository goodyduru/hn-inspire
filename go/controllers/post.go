package controllers

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/goodyduru/go-news/models"
	"github.com/goodyduru/go-news/sessions"
)

func submitForm(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	sess := ctx.Value(hnContextKey("sess")).(sessions.Session)
	p := pageData{User: sess.Get("user").(*models.User)}
	f := form{Token: generateToken()}
	sess.Set("token", f.Token)
	p.Form = f
	if err := renderTemplate(w, "submit", "base.html", &p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func submit(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	sess := ctx.Value(hnContextKey("sess")).(sessions.Session)
	storedToken := sess.Get("token")
	if storedToken == nil {
		http.Error(w, "Invalid submission", http.StatusForbidden)
		return
	}
	user := sess.Get("user").(*models.User)
	storedToken = storedToken.(string)
	token := r.PostFormValue("token")
	title := strings.TrimSpace(r.PostFormValue("title"))
	postUrl := strings.TrimSpace(r.PostFormValue("url"))
	text := strings.TrimSpace(r.PostFormValue("text"))
	f := form{Token: generateToken()}
	sess.Set("token", f.Token)
	p := pageData{User: user, Form: f}

	formErrors := make([]string, 0)
	if token != storedToken {
		formErrors = append(formErrors, "Invalid token")
	}
	if title == "" {
		formErrors = append(formErrors, "Empty title is not allowed")
	}
	if postUrl == "" && text == "" {
		formErrors = append(formErrors, "URL and Text cannot be empty at the same time")
	}

	if _, err := url.ParseRequestURI(postUrl); postUrl != "" && err != nil {
		formErrors = append(formErrors, "Invalid URL")
	}
	if len(formErrors) > 0 {
		p.Errors = formErrors
		if err := renderTemplate(w, "submit", "base.html", &p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	post := models.Post{
		Title:    title,
		Url:      postUrl,
		Text:     text,
		AuthorID: user.ID,
	}
	if err := post.Create(); err != nil {
		formErrors = append(formErrors, err.Error())
		p.Errors = formErrors
		if err := renderTemplate(w, "submit", "base.html", &p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	sess.Delete("token")

	http.Redirect(w, r, "/", http.StatusFound)
}

func all(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	sess := ctx.Value(hnContextKey("sess")).(sessions.Session)
	user := sess.Get("user")
	p := &pageData{}
	var posts []models.Post
	var err error
	if user != nil {
		p.User = user.(*models.User)
		posts, err = models.GetAuthRanked(p.User.ID)
	} else {
		posts, err = models.GetAuthRanked(8)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	p.Posts = posts
	renderTemplate(w, "index", "base.html", p)
}
