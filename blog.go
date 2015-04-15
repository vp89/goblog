package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// Config struct used for database connection
type Config struct {
	Server       string
	Username     string
	Password     string
	Database     string
	CookieSecret string
}

// Post struct used for blog post result set
type Post struct {
	ID    int
	Title string
	// formatted with dashes instead of spaces
	TitleLink  string
	Body       string
	CreateDate time.Time `db:"create_date"`
	// formatted for front page
	CreateDateFmt string
	ModifyDate    time.Time `db:"modify_date"`
	ModifyDateFmt string
}

var store sessions.CookieStore

var db *sqlx.DB
var err error
var conf Config

var templateFuncMap = template.FuncMap{
	"markDown": markDowner,
}

var templates = template.Must(template.New("").Funcs(templateFuncMap).ParseGlob("templates/*"))

func main() {
	unpackConfig()
	connectDatabase()

	store = *sessions.NewCookieStore([]byte(conf.CookieSecret))

	r := mux.NewRouter().StrictSlash(false)

	// serve static resources file
	var resourcesPath = flag.String("resources", "resources/", "Path to resources files")
	r.PathPrefix("/resources/").Handler(http.StripPrefix("/resources/", http.FileServer(http.Dir(*resourcesPath))))

	r.HandleFunc("/admin/delete/{id}", deletePost)
	r.HandleFunc("/admin/edit/{id}", editPost)
	r.HandleFunc("/admin", getAdmin)
	r.HandleFunc("/", getPostTitles)

	// view a single post by title
	post := r.PathPrefix("/{title}").Subrouter()
	post.Methods("GET").HandlerFunc(getPost)

	http.ListenAndServe(":3000", r)
}

func markDowner(args ...interface{}) template.HTML {
	s := blackfriday.MarkdownCommon([]byte(fmt.Sprintf("%s", args...)))
	s = bluemonday.UGCPolicy().SanitizeBytes(s)
	return template.HTML(s)
}

// get all post titles from posts table
func getPostTitles(w http.ResponseWriter, r *http.Request) {
	// populate array of PostTitle from database query
	posts := []Post{}
	_ = db.Select(&posts, "select id, title, create_date, modify_date from posts order by create_date desc")

	// add dashes to post URL
	for key := range posts {
		posts[key].TitleLink = strings.Replace(posts[key].Title, " ", "-", -1)
		posts[key].CreateDateFmt = fmt.Sprintf("%d/%d", posts[key].CreateDate.Month(), posts[key].CreateDate.Day())
	}

	// write response using template file and array of PostTitle
	_ = templates.ExecuteTemplate(w, "index", posts)
}

// get a single post and display it
func getPost(w http.ResponseWriter, r *http.Request) {
	// get the title from URL, replace dashes for spaces for database query
	title := strings.Replace(mux.Vars(r)["title"], "-", " ", -1)

	post := Post{}
	db.Get(&post, "select id, title, body, create_date, modify_date from posts where title = $1", title)

	_ = templates.ExecuteTemplate(w, "post", post)
}

func authenticateAdmin(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
	if r.Method == "GET" {
		_ = templates.ExecuteTemplate(w, "admin_login", nil)
	} else {
		user := r.FormValue("login-user")
		password := r.FormValue("login-password")
		a, _ := db.Exec("select * from users where user_name = $1 and password = $2", user, password)
		b, _ := a.RowsAffected()
		if b > 0 {
			session.Values["logged_in"] = "true"
			session.Save(r, w)
			http.Redirect(w, r, "/admin", 301)
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
			_ = db.Select(&posts, "select id, title, create_date, modify_date from posts order by create_date desc")
			for key := range posts {
				cd := posts[key].CreateDate
				md := posts[key].ModifyDate
				posts[key].CreateDateFmt = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", cd.Year(), cd.Month(), cd.Day(), cd.Hour(), cd.Minute(), cd.Second())
				posts[key].ModifyDateFmt = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", md.Year(), md.Month(), md.Day(), md.Hour(), md.Minute(), md.Second())
			}
			_ = templates.ExecuteTemplate(w, "admin", posts)
			// on POST, insert to database
		} else {
			title := r.FormValue("post-title")
			body := r.FormValue("post-body")

			_, err = db.Exec("insert into posts (title, body, create_date, modify_date) values($1, $2, $3, $4)", title, body, time.Now(), time.Now())
			checkErr(err)
		}
	}
}

func deletePost(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	db.Exec("delete from posts where id = $1", id)
	http.Redirect(w, r, "/admin", 301)
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
			_ = templates.ExecuteTemplate(w, "admin_edit", post)

		} else {
			title := r.FormValue("post-title")
			body := r.FormValue("post-body")
			_, err = db.Exec("update posts set title = $1, body = $2, modify_date = $3 where id = $4", title, body, time.Now(), id)
			checkErr(err)
			http.Redirect(w, r, "/admin/edit/"+id, 301)
		}
	}
}

func connectDatabase() {
	db, err = sqlx.Open("postgres", "postgres://"+conf.Username+":"+conf.Password+"@"+conf.Server+"/"+conf.Database)
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func unpackConfig() {
	content, _ := ioutil.ReadFile("config.json")
	err := json.Unmarshal(content, &conf)
	checkErr(err)
}
