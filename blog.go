package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gopkg.in/fsnotify.v1"
)

// Config struct used for database connection
type Config struct {
	Server, Username, Password, Database, CookieSecret string
}
var (
	store sessions.CookieStore
	db *sqlx.DB
	err error
	conf Config
)

var templateFuncMap = template.FuncMap{
	"markDown":          markDowner,
	"titleLink":         titleLinker,
	"dateFormat":        dateFormatter,
	"dateFormatNice":    dateFormatterNice,
	"draftText":         draftText,
	"draftClass":        draftClass,
	"dateFormatWorkout": dateFormatterWorkouts,
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

	r.HandleFunc("/Publish/{id}", publishPost)
	r.HandleFunc("/Unpublish/{id}", unpublishPost)
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

func connectDatabase() {
	db, err = sqlx.Open("postgres", "postgres://"+conf.Username+":"+conf.Password+"@"+conf.Server+"/"+conf.Database)
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

func unpackConfig() {
	content, err2 := ioutil.ReadFile("config.json")
	checkErr(err2)
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
