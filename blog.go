package main

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gopkg.in/fsnotify.v1"
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
	ID         int
	Title      string
	Body       string
	CreateDate time.Time `db:"create_date"`
	ModifyDate time.Time `db:"modify_date"`
}

var store sessions.CookieStore

var db *sqlx.DB
var err error
var conf Config

var templateFuncMap = template.FuncMap{
	"markDown":       markDowner,
	"titleLink":      titleLinker,
	"dateFormat":     dateFormatter,
	"dateFormatNice": dateFormatterNice,
}

var templates = template.Must(template.New("").Funcs(templateFuncMap).ParseGlob("templates/*"))

func main() {
	unpackConfig()
	connectDatabase()
	go startTemplateRefresher()

	store = *sessions.NewCookieStore([]byte(conf.CookieSecret))

	r := mux.NewRouter().StrictSlash(false)

	// serve static resources file
	var resourcesPath = flag.String("resources", "resources/", "Path to resources files")
	r.PathPrefix("/resources/").Handler(http.StripPrefix("/resources/", http.FileServer(http.Dir(*resourcesPath))))

	r.HandleFunc("/markdown", getMarkdownPreview)
	r.HandleFunc("/admin/new", getAdminNewPost)
	r.HandleFunc("/admin/delete/{id}", deletePost)
	r.HandleFunc("/admin/edit/{id}", editPost)
	r.HandleFunc("/admin", getAdmin)
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
	_ = db.Select(&posts, "select id, title, create_date, modify_date from posts order by create_date desc")
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
			_ = templates.ExecuteTemplate(w, "admin", posts)
			// on POST, insert to database
		}
	}
}

func getAdminNewPost(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		_ = templates.ExecuteTemplate(w, "admin_new", nil)
		// on POST, insert to database
	} else {
		title := r.FormValue("post-title")
		body := r.FormValue("post-body")
		db.Exec("insert into posts (title, body, create_date, modify_date) values($1, $2, $3, $4)", title, body, time.Now(), time.Now())
		http.Redirect(w, r, "/admin", 301)
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
			templates.ExecuteTemplate(w, "admin_edit", post)

		} else {
			title := r.FormValue("post-title")
			body := r.FormValue("post-body")
			db.Exec("update posts set title = $1, body = $2, modify_date = $3 where id = $4", title, body, time.Now(), id)
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

// this allows you to effect UI changes through templates
// without restarting the app
func startTemplateRefresher() {
	watcher, err := fsnotify.NewWatcher()
	checkErr(err)

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					templates = template.Must(template.New("").Funcs(templateFuncMap).ParseGlob("templates/*"))
				}
			case err := <-watcher.Errors:
				checkErr(err)
			}
		}
	}()

	err = watcher.Add("templates")
	checkErr(err)

	<-done
}
