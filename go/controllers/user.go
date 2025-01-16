package controllers

import (
	"net/http"
	"net/mail"
	"strings"

	"github.com/goodyduru/go-news/models"
	"github.com/goodyduru/go-news/sessions"
)

func profile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess := ctx.Value(hnContextKey("sess")).(sessions.Session)
	u := ctx.Value(hnContextKey("user")).(*models.User)
	currentUser := sess.Get("user")
	p := &pageData{User: u}
	if currentUser != nil {
		p.CurrentUser = currentUser.(*models.User)
		if p.CurrentUser.ID == u.ID {
			p.Form.Token = generateToken()
			sess.Set("token", p.Form.Token)
		}
	}
	renderTemplate(w, "user", p)
}

func updateProfile(w http.ResponseWriter, r *http.Request) {
	sess := r.Context().Value(hnContextKey("sess")).(sessions.Session)
	storedToken := sess.Get("token")
	if storedToken == nil {
		http.Error(w, "Invalid submission", http.StatusForbidden)
		return
	}
	user := sess.Get("user").(*models.User)
	storedToken = storedToken.(string)
	token := r.PostFormValue(("token"))
	username := strings.TrimSpace(r.PostFormValue("id"))
	email := strings.TrimSpace(r.PostFormValue("email"))
	var emailErr error
	if email != "" {
		_, emailErr = mail.ParseAddress(email)
	}
	about := strings.TrimSpace(r.PostFormValue("about"))
	if username != user.Username || emailErr != nil || token != storedToken {
		http.Error(w, "Invalid submission", http.StatusForbidden)
		return
	}
	user.Email = email
	user.About = about
	if err := user.Update(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/user?id="+user.Username, http.StatusFound)
}
