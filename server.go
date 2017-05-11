package main

import (
    "io/ioutil"
    "net/http"
    "crypto/tls"
    "database/sql"
    _ "github.com/lib/pq"
    "regexp"
    "time"
    "log"
    "encoding/json"
)

var (
    db  *sql.DB
)

type Profile struct {
    FirstName   string
    LastName    string
    Email       string
    DOB         time.Time
    Bio         string
}

type Post struct {
    Error       string
    AuthorUrl   string
    Location    string
    Timestamp   time.Time
    Content     []byte
}

var validPath = regexp.MustCompile("^/(home|profile|data)/([a-zA-Z0-9]+)$")

func loadProfile(url string) (*Profile, error) {
    var (
        firstname   string
        lastname    string
        email       string
        dob         time.Time
        bio         string
    )

    err := db.QueryRow("SELECT firstname, lastname, email, dob, bio FROM profile WHERE url = $1", url).Scan(&firstname, &lastname, &email, &dob, &bio)
    if err != nil {
        return nil, err
    }

    return &Profile {FirstName: firstname,
                    LastName: lastname,
                    Email: email,
                    DOB: dob,
                    Bio: bio}, nil
}

func profileHandler(w http.ResponseWriter, r *http.Request, url string) {
    log.Println("Profile " + url + " requested.");

    addHeaders(w, r)

    p, err := loadProfile(url)

    if err != nil {
        if err != sql.ErrNoRows {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            log.Println(err)
            return
        } else {
            p = &Profile {FirstName: "User does not exist.", LastName: "", Email: "", DOB: time.Time{}, Bio: ""}
        }
    }

    err = json.NewEncoder(w).Encode(p)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }

    log.Println("Profile " + url + " encoded.")
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            log.Println(r.URL.Path + " not found.")
            return
        }
        fn(w, r, m[2])
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
    cer, err := tls.LoadX509KeyPair(cPath + "fullchain.pem", cPath + "privkey.pem")

    if err != nil {
        log.Fatal(err)
    }

    config := &tls.Config{Certificates: []tls.Certificate{cer}}
    ln, err := tls.Listen("tcp", ":8080", config)

    if err != nil {
        log.Fatal(err)
    }

    defer ln.Close()

    //r.HandleFunc("/home/", makeHandler(homeHandler))
    http.HandleFunc("/profile/", makeHandler(profileHandler))//.Methods("GET")
    //r.HandleFunc("/data/", makeHandler(dataHandler))

    log.Println("Listening on port 8000...")
    log.Fatal(http.ListenAndServeTLS(":8000", cPath + "fullchain.pem", cPath + "privkey.pem", nil))
}
