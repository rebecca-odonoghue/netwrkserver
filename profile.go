package main

import (
    "net/http"
    "database/sql"
    _ "github.com/lib/pq"
    "log"
    "encoding/json"
    "time"
)

type Profile struct {
    firstname   string
    lastname    string
    email       string
    dob         time.Time
    bio         string
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
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }

    err = json.NewEncoder(w).Encode(p)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }

    log.Println("Profile " + url + " encoded.")
}

func loadProfile(url string) (*Profile, error) {
    var (
        firstname   string
        lastname    string
        email       string
        dob         time.Time
        bio         string
    )

    query := `SELECT firstname, lastname, email, dob, bio
            FROM profile
            WHERE url = $1;`
    err := db.QueryRow(query, url).Scan(&firstname, &lastname, &email, &dob, &bio)
    if err != nil {
        return nil, err
    }

    return &Profile {firstname: firstname,
                    lastname: lastname,
                    email: email,
                    dob: dob,
                    bio: bio}, nil
}

func createProfile(url string, profile *Profile) error {
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

func requestConnection(senderUrl string, recipientUrl string,
        senderDesc string, recipientDesc string) error {
    query := `INSERT INTO connection (from-url, to-url, from-descriptor, to-descriptor)
            VALUES ($1, $2, $3, $4);`

    _, err := db.Exec(query, senderUrl, recipientUrl, senderDesc, recipientDesc)

    return err
}

func acceptConnection(senderUrl string, recipientUrl string) error {
    query := `UPDATE connection
            SET accepted = true
            WHERE from-url = $1 AND to-url = $2;`

    _, err := db.Exec(query, senderUrl, recipientUrl);

    return err
}

func deleteConnection(senderUrl string, recipientUrl string) error {
    query := `DELETE FROM connection
            WHERE from-url = $1 AND to-url = $2;`

    _, err := db.Exec(query, senderUrl, recipientUrl);

    return err
}

