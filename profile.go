package main

import (
    "net/http"
    "database/sql"
    _ "github.com/lib/pq"
    "log"
    "encoding/json"
    "time"
    "regexp"
)

var connectPath = regexp.MustCompile("^(request|accept|delete|modify)/?$")

type Profile struct {
    firstname   string
    lastname    string
    email       string
    dob         time.Time
    bio         string
}

type Connection struct {
    fromUrl     string
    toUrl       string
    fromDesc    string
    toDesc      string
}

func profileHandler(w http.ResponseWriter, r *http.Request, url string) {

    path := editablePath.FindStringSubmatch(url)

    if path == nil {
        http.NotFound(w, r)
        return
    }

    switch path[1] {
    case "get":
        p, err := loadProfile(url)

        if err != nil {
            if err != sql.ErrNoRows {
                http.NotFound(w, r)
            } else {
                http.Error(w, err.Error(), http.StatusInternalServerError)
            }
            log.Println(err)
            return
        }

        err = json.NewEncoder(w).Encode(p)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        log.Println("Profile " + url + " encoded.")
    case "new":
        if r.Body == nil || len(path) < 3 {
            http.Error(w, "Request incomplete", http.StatusBadRequest)
            return
        }

        var p Profile
        err := json.NewDecoder(r.Body).Decode(&p)

        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        err = createProfile(path[2], p)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
    case "delete":
        if len(path) < 3 {
            http.NotFound(w, r)
            return
        }

        var p Profile
        err := json.NewDecoder(r.Body).Decode(&p)

        err = deleteProfile(path[2])

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
    case "modify":
        if r.Body == nil || len(path) < 3 {
            http.Error(w, "Request incomplete", http.StatusBadRequest)
            return
        }

        var p Profile
        err := json.NewDecoder(r.Body).Decode(&p)

        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        err = modifyProfile(path[2], p)

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

func connectionHandler(w http.ResponseWriter, r *http.Request, url string) {

    path := connectPath.FindStringSubmatch(url)

    if path == nil {
        http.NotFound(w, r)
        return
    }

    var c Connection
    err := json.NewDecoder(r.Body).Decode(&c)

    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    switch path[1] {
    case "request":
        err = requestConnection(c)
    case "accept":
        err = acceptConnection(c.fromUrl, c.toUrl)
    case "delete":
        err = deleteConnection(c.fromUrl, c.toUrl)
    case "modify":
        err = modifyConnection(c)
    default:
        http.NotFound(w, r)
    }

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func checkUrlHandler(w http.ResponseWriter, r *http.Request, url string) {

    query := `SELECT *
              FROM profile
              WHERE url = $1;`
    
    err := db.QueryRow(query, url).Scan()

    if err != nil {
        if err == sql.ErrNoRows {
            http.NotFound(w, r)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusOK)
}

func friendListHandler(w http.ResponseWriter, r *http.Request, url string) {
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
    err := row.Scan(&(p.firstname), &(p.lastname), &(p.email), &(p.dob), &(p.bio))

    if err != nil {
        return nil, err
    }

    return &p, nil
}

func loadFriends(userUrl string) (map[string]Profile, error) {
    var friends map[string]Profile

    query := `SELECT friend.firstname, friend.lastname, friend.email, friend.dob, friend.bio
            FROM profile user, profile friend, connection
            WHERE user.url = $1
            AND user.url IN(connection.fromurl, connection.tourl)
            AND friend.url IN(connection.fromurl, connetion.tourl)
            AND user.url != friend.url;`

    rows, err := db.Query(query, userUrl)

    if err != nil {
        return nil, err
    }

    friends = make(map[string]Profile)

    for rows.Next() {
        var p Profile
        err = rows.Scan(&(p.firstname), &(p.lastname), &(p.email), &(p.dob), &(p.bio))
        friends[userUrl] = p
    }

    if err != nil {
        return nil, err
    }

    return friends, nil
}

func createProfile(url string, profile Profile) error {
    query := `INSERT INTO profile (url, firstname, lastname, email, dob, bio)
            VALUES ($1,$2, $3, $4, $5, $6);`

    _, err := db.Exec(query, url, profile.firstname, profile.lastname,
            profile.email, profile.dob, profile.bio)

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

    _, err := db.Exec(query, profile.firstname, profile.lastname,
            profile.dob, profile.bio, url)

    return err
}

func requestConnection(con Connection) error {
    query := `INSERT INTO connection (fromurl, tourl, fromdescriptor, todescriptor)
            VALUES ($1, $2, $3, $4);`

    _, err := db.Exec(query, con.fromUrl, con.toUrl, con.fromDesc, con.toDesc)

    return err
}

func acceptConnection(senderUrl string, recipientUrl string) error {
    query := `UPDATE connection
            SET accepted = true
            WHERE fromurl = $1 AND tourl = $2;`

    _, err := db.Exec(query, senderUrl, recipientUrl);

    return err
}

func deleteConnection(senderUrl string, recipientUrl string) error {
    query := `DELETE FROM connection
            WHERE fromurl = $1 AND tourl = $2;`

    _, err := db.Exec(query, senderUrl, recipientUrl);

    return err
}

func modifyConnection(con Connection) error {
    query := `UPDATE connection
            SET fromdescriptor = $1, todescriptor = $2
            WHERE fromurl = $3 
            AND tourl = $4;`

    _, err := db.Exec(query, con.fromDesc, con.toDesc, con.fromUrl, con.toUrl)

    return err
}
