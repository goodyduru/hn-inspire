package controllers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/goodyduru/go-news/models"
	"github.com/goodyduru/go-news/sessions"
)

func defaultHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var sess sessions.Session
		if ctx.Value(hnContextKey("sess")) == nil {
			sess = sessions.GlobalSessions.SessionStart(w, r)
		} else {
			sess = ctx.Value(hnContextKey("sess")).(sessions.Session)
		}
		ctx = context.WithValue(ctx, hnContextKey("sess"), sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func authHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var sess sessions.Session
		ctx := r.Context()
		if ctx.Value(hnContextKey("sess")) == nil {
			sess = sessions.GlobalSessions.SessionStart(w, r)
		} else {
			sess = ctx.Value(hnContextKey("sess")).(sessions.Session)
		}
		registerToken := sess.Get("register_token")
		loginToken := sess.Get("login_token")
		if registerToken == nil || loginToken == nil {
			http.Error(w, "Invalid submission", http.StatusForbidden)
			return
		}
		var token string
		if strings.Contains(r.URL.Path, "login") {
			token = loginToken.(string)
		} else {
			token = registerToken.(string)
		}
		pageData := loginPage{
			Login:    generateToken(),
			Register: generateToken(),
		}
		sess.Set("login_token", pageData.Login)
		sess.Set("register_token", pageData.Register)
		ctx = context.WithValue(ctx, hnContextKey("sess"), sess)
		ctx = context.WithValue(ctx, hnContextKey("token"), token)
		ctx = context.WithValue(ctx, hnContextKey("page_data"), &pageData)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func loginRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var sess sessions.Session
		if ctx.Value(hnContextKey("sess")) == nil {
			sess = sessions.GlobalSessions.SessionStart(w, r)
		} else {
			sess = ctx.Value(hnContextKey("sess")).(sessions.Session)
		}
		user := sess.Get("user")
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		ctx = context.WithValue(ctx, hnContextKey("sess"), sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func checkUserID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		var sess sessions.Session
		if ctx.Value(hnContextKey("sess")) == nil {
			sess = sessions.GlobalSessions.SessionStart(w, r)
		} else {
			sess = ctx.Value(hnContextKey("sess")).(sessions.Session)
		}
		ctx = context.WithValue(ctx, hnContextKey("sess"), sess)
		ctx = context.WithValue(ctx, hnContextKey("user"), &u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func checkItemID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		post := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(post, 10, 64)
		if err != nil {
			http.Error(w, "No such post exist", http.StatusNotFound)
		}
		var sess sessions.Session
		ctx := r.Context()
		if ctx.Value(hnContextKey("sess")) == nil {
			sess = sessions.GlobalSessions.SessionStart(w, r)
		} else {
			sess = ctx.Value(hnContextKey("sess")).(sessions.Session)
		}
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
		ctx = context.WithValue(ctx, hnContextKey("sess"), sess)
		ctx = context.WithValue(ctx, hnContextKey("post"), p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
