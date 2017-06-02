package main

import (
    "net/http"
    "github.com/gorilla/mux"
    "database/sql"
    _ "github.com/lib/pq"
    "log"
    "encoding/json"
    "time"
    "regexp"
)

var connectPath = regexp.MustCompile("^(request|accept|delete|modify)/?$")

type Profile struct {
    FirstName   string      `json:"firstname"`
    LastName    string      `json:"lastname"`
    Email       string      `json:"email"`
    DOB         time.Time   `json:"dob"`
    Bio         string      `json:"bio"`
}

type Connection struct {
    FromUrl     string      `json:"fromUrl"`
    ToUrl       string      `json:"toUrl"`
    FromDesc    string      `json:"fromDesc"`
    ToDesc      string      `json:"toDesc"`
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    action := vars["action"]
    url := vars["url"]

    switch action {
    case "get":
        p, err := loadProfile(url)

        if err != nil {
            if err == sql.ErrNoRows {
                http.NotFound(w, r)
            } else {
                http.Error(w, err.Error(), http.StatusInternalServerError)
            }
            log.Println(err)
            return
        }

        err = json.NewEncoder(w).Encode(&p)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        log.Println("Profile " + url + " encoded.")
    case "new":
        if r.Body == nil || url == "" {
            http.Error(w, "Request incomplete", http.StatusBadRequest)
            return
        }

        var p Profile
        err := json.NewDecoder(r.Body).Decode(&p)

        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        err = createProfile(url, p)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
    case "delete":
        if url == "" {
            http.NotFound(w, r)
            return
        }

        var p Profile
        err := json.NewDecoder(r.Body).Decode(&p)

        err = deleteProfile(url)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
    case "modify":
        if r.Body == nil || url == "" {
            http.Error(w, "Request incomplete", http.StatusBadRequest)
            return
        }

        var p Profile
        err := json.NewDecoder(r.Body).Decode(&p)

        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        err = modifyProfile(url, p)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
    default:
        http.NotFound(w, r)
        return
    }

}

func connectionHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    action := vars["action"]
    p1 := vars["p1"]
    p2 := vars["p2"]

    //var c Connection
    //err := json.NewDecoder(r.Body).Decode(&c)

    //if err != nil {
      //  http.Error(w, err.Error(), http.StatusBadRequest)
       // return
   // }
    var err error
    switch action {
    case "get":
        log.Println("connection check!")
        exists, accepted, requestedBy := connectionExists(p1,p2)
        
        if exists {
            log.Println("exists")
            log.Println("requestedBy: " + requestedBy);
        }
        if accepted {
            log.Println("accepted")
        }
        err = json.NewEncoder(w).Encode(struct {
            Exists bool `json:"exists"`
            Accepted bool `json:"accepted"`
            RequestedBy string `json:"requestedBy"`
        }{
            Exists: exists,
            Accepted: accepted,
            RequestedBy: requestedBy,
        })

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    case "request":
        err = requestConnection(p1, p2)
    case "accept":
        err = acceptConnection(p1, p2)
    case "delete":
        err = deleteConnection(p1, p2)
    case "modify":
        err = modifyConnection(p1, p2)
    default:
        http.NotFound(w, r)
        return
    }

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }

}

func checkUrlHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    url := vars["url"]

    query := `SELECT url 
              FROM profile
              WHERE url = $1;`

    err := db.QueryRow(query, url).Scan(&url)
    var available = false

    if err != nil {
        if err == sql.ErrNoRows {
            available = true
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }

    err = json.NewEncoder(w).Encode(
        struct {
            Available bool `json:"available"`
        }{
            Available: available,
        })

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func friendListHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    url := vars["url"]

    friends, err := loadFriends(url)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    err = json.NewEncoder(w).Encode(friends)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

func loadProfile(url string) (*Profile, error) {

    query := `SELECT firstname, lastname, email, dob, bio
            FROM profile
            WHERE url = $1;`

    var p Profile
    row := db.QueryRow(query, url)
    err := row.Scan(&(p.FirstName), &(p.LastName), &(p.Email), &(p.DOB), &(p.Bio))

    if err != nil {
        return nil, err
    }

    return &p, nil
}

func loadFriends(userUrl string) ([]struct{
        URL string  `json:"url"`
        P Profile `json:"profile"`
    }, error) {

    var friends []struct{
        URL string  `json:"url"`
        P Profile `json:"profile"`
    }

    query := `SELECT friend.url, friend.firstname, friend.lastname, friend.email, friend.dob, friend.bio
            FROM profile user, profile friend, connection c
            WHERE user.url = $1
            AND user.url IN(c.fromurl, c.tourl)
            AND friend.url IN(c.fromurl, c.tourl)
            AND user.url <> friend.url;`

    rows, err := db.Query(query, userUrl)

    if err != nil {
        return nil, err
    }

    for rows.Next() {
        var friend struct{
            URL string  `json:"url"`
            P Profile `json:"profile"`
        }

        err = rows.Scan(&friend.URL,
                &friend.P.FirstName,
                &friend.P.LastName,
                &friend.P.Email,
                &friend.P.DOB,
                &friend.P.Bio)
        friends = append(friends, friend)
    }

    if err != nil {
        return nil, err
    }

    return friends, nil
}

func createProfile(url string, profile Profile) error {
    query := `INSERT INTO profile (url, firstname, lastname, email, dob, bio)
            VALUES ($1,$2, $3, $4, $5, $6);`

    _, err := db.Exec(query, url, profile.FirstName, profile.LastName,
            profile.Email, profile.DOB, profile.Bio)

    return err
}

func deleteProfile(url string) error {
    query := `DELETE FROM profile
            WHERE url = $1;`

    _, err := db.Exec(query, url)

    return err
}

func modifyProfile(url string, profile Profile) error {

    query := `UPDATE profile
            SET firstname = $1, lastname = $2, dob = $3, bio = $4
            WHERE url = $5;`

    _, err := db.Exec(query, profile.FirstName, profile.LastName,
            profile.DOB, profile.Bio, url)

    return err
}

func connectionExists(p1 string, p2 string) (bool, bool, string) {

    query := `SELECT c.accepted, c.fromurl
            FROM connection c
            WHERE $1 IN(c.fromurl, c.tourl)
            AND $2 IN(c.fromurl, c.tourl);`

    var accepted bool
    var requestedBy string

    err := db.QueryRow(query, p1, p2).Scan(&accepted, &requestedBy)

    if err != nil {
        log.Println(err.Error());
        return false, false, ""
    }

    return true, accepted, requestedBy
}

func requestConnection(p1 string, p2 string) error {
    query := `INSERT INTO connection (fromurl, tourl, fromdescriptor, todescriptor)
            VALUES ($1, $2, $3, $4);`

    _, err := db.Exec(query, p1, p2, "friend", "friend")

    return err
}

func acceptConnection(p1 string, p2 string) error {
    query := `UPDATE connection 
            SET accepted = true
            WHERE fromurl IN($1, $2)
            AND tourl IN($1, $2);`

    _, err := db.Exec(query, p1, p2);

    return err
}

func deleteConnection(p1 string, p2 string) error {
    query := `DELETE FROM connection
            WHERE fromurl IN($1, $2)
            AND tourl IN($1, $2);`

    _, err := db.Exec(query, p1, p2);

    return err
}

func modifyConnection(p1 string, p2 string) error {
    query := `UPDATE connection
            SET fromdescriptor = $1, todescriptor = $2
            WHERE fromurl = $3 
            AND tourl = $4;`

    _, err := db.Exec(query, "friend", "friend", p1, p2)

    return err
}
