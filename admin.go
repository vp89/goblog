package main

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

func authenticateAdmin(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
	if r.Method == "GET" {
		err = templates.ExecuteTemplate(w, "admin_login", nil)
		checkErr(err)
	} else {
		user := r.FormValue("login-user")
		password := r.FormValue("login-password")
		a, err2 := db.Exec("select * from users where user_name = $1 and password = $2", user, password)
		checkErr(err2)
		b, err2 := a.RowsAffected()
		checkErr(err2)
		if b > 0 {
			session.Values["logged_in"] = "true"
			session.Save(r, w)
			http.Redirect(w, r, "/admin", 302)
		}
	}
}

func getAdmin(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "blog_admin")
	if session.Values["logged_in"] == nil {
		authenticateAdmin(w, r, session)
	} else {
		if r.Method == "GET" {
			// populate array of PostTitle from database query
			posts := []Post{}
			err = db.Select(&posts, "select id, title, create_date, modify_date, draft from posts order by create_date desc")
			checkErr(err)
			err = templates.ExecuteTemplate(w, "admin", posts)
			checkErr(err)
			// on POST, insert to database
		}
	}
}

func getAdminNewPost(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "blog_admin")
	if session.Values["logged_in"] == nil {
		authenticateAdmin(w, r, session)
	} else {
		if r.Method == "GET" {
			err = templates.ExecuteTemplate(w, "admin_new", nil)
			checkErr(err)
			// on POST, insert to database
		} else {
			title := r.FormValue("post-title")
			body := r.FormValue("post-body")
			db.Exec("insert into posts (title, body, create_date, modify_date, draft) values($1, $2, $3, $4, 1)", title, body, time.Now(), time.Now())
			http.Redirect(w, r, "/admin", 302)
		}
	}
}
