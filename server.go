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
    "encoding/json"
    "github.com/gorilla/mux"
)

type Server struct {
    r *mux.Router
}

type Profile struct {
    //Url     string
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
    fmt.Println("Profile " + url + " requested.");
    serveHTTP(w, r)
    p, err := loadProfile(url)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        fmt.Println(err)
        return
    }
    err = json.NewEncoder(w).Encode(p)
    fmt.Println("Profile " + url + " encoded.")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        fmt.Println(err)
    }
}

func dataHandler(w http.ResponseWriter, r *http.Request, url string) {}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    fmt.Println("Making Handler...");
    return func(w http.ResponseWriter, r *http.Request) {
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            fmt.Println(r.URL.Path + " not found.")
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

func serveHTTP(w http.ResponseWriter, r *http.Request) {
    if origin := r.Header.Get("Origin"); origin != "" {
        w.Header().Set("Access-Control-Allow-Origin", origin)
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers",
            "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
    }

    // Stop here if its Preflighted OPTIONS request
    if r.Method == "OPTIONS" {
        return
    }

    fmt.Println("Serving HTTP headers.")
    // Lets Gorilla work
    //s.r.ServeHTTP(w, r)
}

func main() {
    var err error
    db, err = sql.Open("postgres", "user=postgres password=Mond@y4.55 dbname=opennetwork")

    if err != nil {
        log.Fatal(err)
    }

    defer db.Close()

    //r.HandleFunc("/home/", makeHandler(homeHandler))
    http.HandleFunc("/profile/", makeHandler(profileHandler))//.Methods("GET")
    //r.HandleFunc("/data/", makeHandler(dataHandler))
    
    fmt.Println("Listening on port 8080...")
    http.ListenAndServe(":8080", nil)
}
