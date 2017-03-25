package main

import (
    "net/http"
    "database/sql"
    _ "github.com/lib/pq"
    "regexp"
    "time"
    "html/template"
    "log"
    "fmt"
)

type Profile struct {
    Name    string
    Motto   string
}

type Post struct {
    Author      string
    Location    string
    Timestamp   time.Time
    Content     []byte
}

var db *sql.DB
var templates = template.Must(template.ParseFiles("profile.html"))
var validPath = regexp.MustCompile("^/(home|profile|data)/([a-zA-Z0-9]+)$")

func loadProfile(url string) (*Profile, error) {
    var (
        name string
        motto string
    )

    err := db.QueryRow("SELECT name, motto FROM profile WHERE url = $1", url).Scan(&name, &motto)
    if err != nil {
        return nil, err
    }

    return &Profile{Name: name, Motto: motto}, nil
}

func homeHandler(w http.ResponseWriter, r *http.Request, url string) {}

func profileHandler(w http.ResponseWriter, r *http.Request, url string) {
    p, err := loadProfile(url)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    renderTemplate(w, "profile", p)
}

func dataHandler(w http.ResponseWriter, r *http.Request, url string) {}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }
        fmt.Printf("%q\n", m)
        fn(w, r, m[2])
    }
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Profile) {
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func main() {
    var err error
    db, err = sql.Open("postgres", "user=postgres password=Mond@y4.55 dbname=opennetwork")

    if err != nil {
        log.Fatal(err)
    }

    defer db.Close()

    http.HandleFunc("/home/", makeHandler(homeHandler))
    http.HandleFunc("/profile/", makeHandler(profileHandler))
    http.HandleFunc("/data/", makeHandler(dataHandler))
    http.ListenAndServe(":8080", nil)
}
