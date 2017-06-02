package main

import (
    "io/ioutil"
    "net/http"
    "database/sql"
    _ "github.com/lib/pq"
    "regexp"
    "github.com/gorilla/mux"
    "log"
)

var validPath = regexp.MustCompile("^/(profile|check|connect|account|register|post|comment|search|authenticate|feed)(/[a-zA-Z0-9])*$")
var editablePath = regexp.MustCompile("^(get|new|delete|modify)/?([a-zA-Z-1-9]*)$")

var db  *sql.DB
type NetwrkServer struct {
    r *mux.Router
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        log.Println(r.URL.Path + " requested.");

        m := validPath.FindStringSubmatch(r.URL.Path)

        if m == nil {
            http.NotFound(w, r)
            log.Println(r.URL.Path + " not found.")
            return
        }

        path := ""

        if len(m) > 2 {
            path = m[2]
        }

        fn(w, r, path)
    }
}

func (s *NetwrkServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if origin := r.Header.Get("Origin"); origin != "" {
        w.Header().Set("Access-Control-Allow-Origin", origin)
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers",
            "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
    }

    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK);
        return
    }

    s.r.ServeHTTP(w, r)
}

func checkAuthorisation(w http.ResponseWriter, r *http.Request) (string, bool) {
    email, password, ok := r.BasicAuth()

    if !ok {
        w.Header().Set("Authorization", "Basic realm=loggedin")
        w.WriteHeader(401)
        log.Println("Authorisation requested")
        return "", false
    }

    err := authenticate(email, password)

    if err != nil {
        http.Error(w, err.Error(), http.StatusForbidden)
        return "", false
    }

    return email, true
}

func main() {
    // Connect to database
    pwd, err := ioutil.ReadFile("auth")

    if err != nil {
        log.Fatal(err)
    }

    db, err = sql.Open("postgres", "user=postgres password=" + string(pwd) + " dbname=netwrk sslmode=require")

    if err != nil {
        log.Fatal(err)
    }

    defer db.Close()

    // TLS Certificate Path
    cPath := "/etc/letsencrypt/live/netwrk.website/"

    // Request Handler Functions
    r := mux.NewRouter()

    r.HandleFunc("/profile/{action}/{url}", profileHandler)
    r.HandleFunc("/check/{url}", checkUrlHandler)
    r.HandleFunc("/connect/{action}/{p1}/{p2}", connectionHandler)
    r.HandleFunc("/post/{action}/{id}",postHandler)
    r.HandleFunc("/comment{id}", makeHandler(commentHandler))
    r.HandleFunc("/search/recent", makeHandler(recentSearchHandler))
    r.HandleFunc("/search/save/{term}", makeHandler(saveSearchHandler))
    r.HandleFunc("/search/{term}", searchHandler)
    r.HandleFunc("/authenticate", makeHandler(authenticationHandler))
    r.HandleFunc("/register", makeHandler(registrationHandler))
    r.HandleFunc("/account/{action}", accountHandler)
    r.HandleFunc("/friends/{url}", friendListHandler)
    r.HandleFunc("/feed", feedHandler)

    http.Handle("/", &NetwrkServer{r})

    log.Println("Listening on port 8000...")
    log.Fatal(http.ListenAndServeTLS(":8000", cPath + "fullchain.pem", cPath + "privkey.pem", nil))
}
