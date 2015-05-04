package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Post struct used for blog post result set
type Post struct {
	ID int
	Title, Body string
	CreateDate time.Time `db:"create_date"`
	ModifyDate time.Time `db:"modify_date"`
	Draft int `db:"draft"`
}

// get all post titles from posts table
func getPostTitles(w http.ResponseWriter, r *http.Request) {
	// populate array of PostTitle from database query
	posts := []Post{}
	err = db.Select(&posts, "select id, title, create_date, modify_date from posts where draft = 0 order by create_date desc")
	checkErr(err)
	err = templates.ExecuteTemplate(w, "index", posts)
	checkErr(err)
}

// get a single post and display it
func getPost(w http.ResponseWriter, r *http.Request) {
	loggedIn := 0
	session, _ := store.Get(r, "blog_admin")

	if session.Values["logged_in"] != nil {
		loggedIn = 1
	}

	title := strings.Replace(mux.Vars(r)["title"], "-", " ", -1)

	post, blankCheck := Post{}, Post{}

	db.Get(&post, "select id, title, body, create_date, modify_date from posts where title = $1 and draft in (0,$2)", title, loggedIn)

	if post == blankCheck {
		http.Redirect(w, r, "/", 302)
		return
	}

	err = templates.ExecuteTemplate(w, "post", post)
	checkErr(err)
}

func publishPost(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "blog_admin")
	if session.Values["logged_in"] == nil {
		authenticateAdmin(w, r, session)
	} else {
		id := mux.Vars(r)["id"]
		db.Exec("update posts set draft = 0 where id = $1", id)
		http.Redirect(w, r, "/admin", 302)
	}
}

func unpublishPost(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "blog_admin")
	if session.Values["logged_in"] == nil {
		authenticateAdmin(w, r, session)
	} else {
		id := mux.Vars(r)["id"]
		db.Exec("update posts set draft = 1 where id = $1", id)
		http.Redirect(w, r, "/admin", 302)
	}
}

func deletePost(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "blog_admin")
	if session.Values["logged_in"] == nil {
		authenticateAdmin(w, r, session)
	} else {
		id := mux.Vars(r)["id"]
		db.Exec("delete from posts where id = $1", id)
		http.Redirect(w, r, "/admin", 302)
	}
}

func editPost(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	session, _ := store.Get(r, "blog_admin")
	if session.Values["logged_in"] == nil {
		authenticateAdmin(w, r, session)
	} else {
		if r.Method == "GET" {
			post := Post{}
			db.Get(&post, "select id, title, body, create_date, modify_date from posts where id = $1", id)
			templates.ExecuteTemplate(w, "admin_edit", post)

		} else {
			title := r.FormValue("post-title")
			body := r.FormValue("post-body")
			db.Exec("update posts set title = $1, body = $2, modify_date = $3 where id = $4", title, body, time.Now(), id)
			http.Redirect(w, r, "/admin/edit/"+id, 302)
		}
	}
}
