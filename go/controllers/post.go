package controllers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/goodyduru/go-news/models"
	"github.com/goodyduru/go-news/sessions"
)

func submitForm(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	sess := ctx.Value(hnContextKey("sess")).(sessions.Session)
	p := pageData{CurrentUser: sess.Get("user").(*models.User)}
	f := generateToken()
	sess.Set("token", f)
	p.Form.Token = f
	if err := renderTemplate(w, "submit", &p); err != nil {
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
	if _, err := url.ParseRequestURI(postUrl); postUrl != "" && err != nil {
		http.Error(w, "Invalid submission", http.StatusForbidden)
		return
	}
	f := generateToken()
	sess.Set("token", f)
	p := pageData{CurrentUser: user, Form: form{Token: f}}

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
	if len(formErrors) > 0 {
		p.Errors = formErrors
		if err := renderTemplate(w, "submit", &p); err != nil {
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
		if err := renderTemplate(w, "submit", &p); err != nil {
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
		p.CurrentUser = user.(*models.User)
		posts, err = models.GetAuthRanked(p.CurrentUser.ID)
	} else {
		posts, err = models.GetRanked()
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	p.Posts = posts
	p.Errors = append(p.Errors, "No post has been submitted.")
	renderTemplate(w, "index", p)
}

func submitted(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	sess := ctx.Value(hnContextKey("sess")).(sessions.Session)
	u := ctx.Value(hnContextKey("user")).(*models.User)
	currentUser := sess.Get("user")
	p := &pageData{}
	var posts []models.Post
	var err error
	if currentUser != nil {
		p.CurrentUser = currentUser.(*models.User)
		posts, err = models.GetAuthSubmitted(u.ID, p.CurrentUser.ID)
	} else {
		posts, err = models.GetSubmitted(u.ID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for p := range posts {
		posts[p].Author = u.Username
	}
	e := fmt.Sprintf("%s hasn't submitted any post.", u.Username)
	p.Posts = posts
	p.Errors = append(p.Errors, e)
	p.User = u
	p.Title = fmt.Sprintf("%s submissions.", u.Username)
	renderTemplate(w, "index", p)
}

func favorites(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	sess := ctx.Value(hnContextKey("sess")).(sessions.Session)
	u := ctx.Value(hnContextKey("user")).(*models.User)
	currentUser := sess.Get("user")
	p := &pageData{}
	var posts []models.Post
	var err error
	if currentUser != nil {
		p.CurrentUser = currentUser.(*models.User)
		posts, err = models.GetAuthMixed(u.ID, p.CurrentUser.ID)
	} else {
		posts, err = models.GetAuthMixed(u.ID, 8)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	e := fmt.Sprintf("%s hasn't favorited any post.", u.Username)
	p.Posts = posts
	p.Errors = append(p.Errors, e)
	p.User = u
	renderTemplate(w, "mixed", p)
}

func comments(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	sess := ctx.Value(hnContextKey("sess")).(sessions.Session)
	u := ctx.Value(hnContextKey("user")).(*models.User)
	currentUser := sess.Get("user")
	p := &pageData{}
	var posts []models.Post
	var err error
	if currentUser != nil {
		p.CurrentUser = currentUser.(*models.User)
		posts, err = models.GetAuthSubmittedComments(u.ID, p.CurrentUser.ID)
	} else {
		posts, err = models.GetSubmittedComments(u.ID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// The comments are already in chronological order, This piece of code sorts the chunks in
	// reverse. The order within each chunk is maintained. The group will be headlined by a top-level comment.
	sortedPosts := make([]models.Post, len(posts))
	prev := len(posts)
	t := 0
	for i := len(posts) - 1; i >= 0; i-- {
		if posts[i].Lev == 0 {
			for j := i; j < prev; j++ {
				sortedPosts[t] = posts[j]
				t++
			}
			prev = i
		}
	}
	addClasses(sortedPosts)
	e := fmt.Sprintf("%s hasn't made any comments.", u.Username)
	p.Posts = posts
	p.Errors = append(p.Errors, e)
	p.User = u
	renderTemplate(w, "threads", p)
}

func addClasses(posts []models.Post) {
	prev_level := 0
	stack := make([]int, 0)
	num_replies := make([]int, 0)
	comment_indices := make([]int, 0)
	ids := make([]int, 0)
	for i := range posts {
		total := 0
		if posts[i].Lev <= prev_level {
			for len(stack) > 0 && stack[len(stack)-1] >= posts[i].Lev {
				ids = ids[:len(ids)-1]
				stack = stack[:len(stack)-1]
				j := comment_indices[len(comment_indices)-1]
				comment_indices = comment_indices[:len(comment_indices)-1]
				total += num_replies[len(num_replies)-1]
				num_replies = num_replies[:len(num_replies)-1]
				posts[j].NumReplies = total
			}
		}
		if len(ids) > 0 {
			num_replies[len(num_replies)-1] += total
			posts[i].HtmlClass = strings.Trim(fmt.Sprint(ids), "[]")
		}
		stack = append(stack, posts[i].Lev)
		ids = append(ids, posts[i].ID)
		num_replies = append(num_replies, 1)
		comment_indices = append(comment_indices, i)
		prev_level = posts[i].Lev
	}
	total := 0
	for len(stack) > 0 {
		stack = stack[:len(stack)-1]
		j := comment_indices[len(comment_indices)-1]
		comment_indices = comment_indices[:len(comment_indices)-1]
		total += num_replies[len(num_replies)-1]
		num_replies = num_replies[:len(num_replies)-1]
		posts[j].NumReplies = total
	}
}

func single(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	sess := ctx.Value(hnContextKey("sess")).(sessions.Session)
	post := ctx.Value(hnContextKey("post")).(models.Post)
	user := sess.Get("user")
	p := &pageData{}
	posts := []models.Post{post}
	var err error
	var comments []models.Post
	if user != nil {
		p.CurrentUser = user.(*models.User)
		token := generateToken()
		sess.Set("token", token)
		p.Form.Token = token
		comments, err = post.GetAuthComments(p.CurrentUser.ID)
	} else {
		comments, err = post.GetComments()
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	addClasses(comments)
	posts = append(posts, comments...)
	p.Posts = posts
	renderTemplate(w, "single", p)
}

func reply(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	sess := ctx.Value(hnContextKey("sess")).(sessions.Session)
	storedToken := sess.Get("token")
	if storedToken == nil {
		http.Error(w, "Invalid submission", http.StatusForbidden)
		return
	}

	user := sess.Get("user").(*models.User)
	storedToken = storedToken.(string)
	parent := r.PostFormValue("parent")
	token := r.PostFormValue("token")
	url := strings.Trim(r.PostFormValue("goto"), "/")
	text := strings.TrimSpace(r.PostFormValue("text"))

	if token != storedToken || text == "" {
		http.Error(w, "Invalid submission", http.StatusForbidden)
		return
	}

	parentID, err := strconv.ParseInt(parent, 10, 64)
	if err != nil {
		http.Error(w, "Invalid submission", http.StatusForbidden)
		return
	}

	p := models.Post{ID: int(parentID)}
	err = p.Get()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if p.AuthorID == 0 {
		http.Error(w, "Invalid submission", http.StatusNotFound)
		return
	}
	newPost := models.Post{
		ParentID: p.ID,
		AuthorID: user.ID,
		Text:     text,
	}
	err = newPost.Create()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sess.Delete("token")
	if !strings.Contains(url, "#") {
		url = fmt.Sprintf("/%s#%d", url, newPost.ID)
	} else {
		url = "/" + url
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func replyForm(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	sess := ctx.Value(hnContextKey("sess")).(sessions.Session)
	user := sess.Get("user").(*models.User)
	parent := r.URL.Query().Get("id")
	url := r.URL.Query().Get("goto")
	parentID, err := strconv.ParseInt(parent, 10, 64)
	if err != nil {
		http.Error(w, "No such post exist", http.StatusNotFound)
		return
	}

	p := models.Post{ID: int(parentID)}
	err = p.GetAuth(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if p.AuthorID == 0 {
		http.Error(w, "No such post exist", http.StatusNotFound)
		return
	}
	if p.ParentID == 0 {
		http.Redirect(w, r, fmt.Sprintf("/item?id=%d", p.ID), http.StatusFound)
		return
	}

	form := form{
		Token: generateToken(),
		Goto:  url,
	}
	posts := []models.Post{p}
	sess.Set("token", form.Token)
	pg := &pageData{CurrentUser: user, Form: form, Posts: posts}
	renderTemplate(w, "reply", pg)
}

func voteOrFlag(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	sess := ctx.Value(hnContextKey("sess")).(sessions.Session)
	post := ctx.Value(hnContextKey("post")).(models.Post)
	user := sess.Get("user")
	var err error
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	u := user.(*models.User)
	path := r.URL.Path
	if path == "/vote" {
		err = post.Vote(u.ID)
	} else if path == "/flag" {
		action := r.URL.Query().Get("action")
		if action == "un" {
			err = post.Unflag(u.ID)
		} else {
			err = post.Flag(u.ID)
		}
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/item?id=%d", post.ID), http.StatusFound)
}
