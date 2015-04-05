package main

import (
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

// Post struct used for blog post result set
type Post struct {
	ID         int
	Title      string
	Body       string
	CreateDate time.Time `db:"create_date"`
	ModifyDate time.Time `db:"modify_date"`
}

var db *sqlx.DB

func main() {
	connectDatabase()

	r := mux.NewRouter().StrictSlash(false)

	// list posts on front page
	r.HandleFunc("/", getPostTitles)

	// view a single post by title
	post := r.PathPrefix("/{title}").Subrouter()
	post.Methods("GET").HandlerFunc(getPost)

	http.ListenAndServe(":3000", r)
}

// get all post titles from posts table
func getPostTitles(w http.ResponseWriter, r *http.Request) {
	// populate array of PostTitle from database query
	posts := []Post{}
	_ = db.Select(&posts, "select id, title, '' as body, create_date, modify_date from posts")

	// write response using template file and array of PostTitle
	tmpl, _ := template.ParseFiles("templates/index.html")
	_ = tmpl.Execute(w, posts)
}

// get a single post and display it
func getPost(w http.ResponseWriter, r *http.Request) {
	// get the title from URL, replace dashes for spaces for database query
	title := strings.Replace(mux.Vars(r)["title"], "-", " ", -1)

	post := Post{}
	db.Get(&post, "select id, title, body, create_date, modify_date from posts where title = $1", title)

	tmpl, _ := template.ParseFiles("templates/post.html")
	_ = tmpl.Execute(w, post)
}

func connectDatabase() {
	var err error
	db, err = sqlx.Open("postgres", "postgres://vince:Arnavisca1@localhost/blog_app")
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
