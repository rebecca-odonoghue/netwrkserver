package main

import (
    "net/http"
    "database/sql"
    _ "github.com/lib/pq"
    "log"
    "encoding/json"
    "time"
)

type Post struct {
    profileUrl  string
    authorUrl   string
    timestamp   time.Time
    content     string
}

func postHandler(w http.ResponseWriter, r *http.Request, id string) {
    log.Println("Post " + id + "requested")

    addHeaders(w, r)

    p, err := loadPost(id)

    if err != nil {
        if err != sql.ErrNoRows {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            log.Println(err)
            return
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            log.Println(err)
            return
        }
    }

    err = json.NewEncoder(w).Encode(p)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        log.Println(err)
    }

}


func loadPost(id string) (*Post, error) {
    var (
        profileUrl  string
        authorUrl   string
        timestamp   time.Time
        content     string
    )

    query := `SELECT profile-url, author-url, timestamp, content
            FROM post
            WHERE id = $1;`

    err := db.QueryRow(query, id).Scan(&profileUrl, &authorUrl, &timestamp, &content)

    if err != nil {
        return nil, err
    }

    return &Post {profileUrl: profileUrl,
            authorUrl: authorUrl,
            timestamp: timestamp,
            content: content}, nil
}

func createPost(id int, post *Post) error {
    query := `INSERT INTO post (profile-url, author-url, timestamp, content)
            VALUES ($1, $2, $3, $4);`

    _, err := db.Exec(query, post.profileUrl, post.authorUrl,
            post.timestamp, post.content)

    return err
}

func deletePost(id int) error {
    query := `DELETE FROM post
            WHERE id = $1;`

    _, err := db.Exec(query, id)

    return err
}

func editPost(id int, content string) error {
    query := `UPDATE post
            SET content = $1
            WHERE id = $2;`

    _, err := db.Exec(query, content, id)

    return err
}
