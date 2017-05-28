package main

import (
    "io/ioutil"
    "net/http"
    "database/sql"
    _ "github.com/lib/pq"
    "regexp"
    "log"
)

var (
    db  *sql.DB
)

var validPath = regexp.MustCompile("^/(profile|post|comment|react|search|authenticate|feed)/([a-zA-Z0-9/]*)$")
/*
func createAccount(email string, dob time.Time) error {
    return db.Query("INSERT INTO account (email, dob) VALUES ($1, $2);", email, dob)
}

func authenticationHandler(w http.ResponseWriter, r *http.Request) {
    
}
*/
func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
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

func addHeaders(w http.ResponseWriter, r *http.Request) {
    if origin := r.Header.Get("Origin"); origin != "" {
        w.Header().Set("Access-Control-Allow-Origin", origin)
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers",
            "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
    }
}

func checkAuthorisation(w http.ResponseWriter, r *http.Request) bool {
    email, password, ok := r.BasicAuth()

    if !ok {
        w.Header().Set("Authorization", "Basic realm=loggedin")
        w.WriteHeader(401)
        log.Println("Authorisation requested")
        return false
    }

    err := authenticate(email, password)

    if err != nil {
        http.Error(w, err.Error(), http.StatusForbidden)
        return false
    }

    return true
}

func main() {
    // Connect to database
    pwd, err := ioutil.ReadFile("auth")

    if err != nil {
        log.Fatal(err)
    }

    db, err = sql.Open("postgres", "user=postgres password=" + string(pwd) + " dbname=netwrk")

    if err != nil {
        log.Fatal(err)
    }

    defer db.Close()

    // TLS
    cPath := "/etc/letsencrypt/live/netwrk.website/"
//    cer, err := tls.LoadX509KeyPair(cPath + "fullchain.pem", cPath + "privkey.pem")

//    if err != nil {
//        log.Fatal(err)
//    }

//    config := &tls.Config{Certificates: []tls.Certificate{cer}}
//    ln, err := tls.Listen("tcp", ":8000", config)

//    if err != nil {
//        log.Fatal(err)
//    }

//    defer ln.Close()

    http.HandleFunc("/profile/", makeHandler(profileHandler))
    http.HandleFunc("/post/", makeHandler(postHandler))
    http.HandleFunc("/comment/", makeHandler(commentHandler))
    http.HandleFunc("/react/", makeHandler(reactionHandler))
    http.HandleFunc("/search/", makeHandler(searchHandler))
    http.HandleFunc("/authenticate/", makeHandler(authenticationHandler))
    http.HandleFunc("/feed/", makeHandler(feedHandler))

    log.Println("Listening on port 8000...")
    log.Fatal(http.ListenAndServeTLS(":8000", cPath + "fullchain.pem", cPath + "privkey.pem", nil))
}
