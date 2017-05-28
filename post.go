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

func postHandler(w http.ResponseWriter, r *http.Request, url string) {

    path := editablePath.FindStringSubmatch(url)

    if path == nil {
        http.NotFound(w,r)
        return
    }

    switch path[1] {
    case "get":
        if len(path) < 3 {
            http.NotFound(w, r)
            return
        }

        p, err := loadPost(path[2])

        if err != nil {
            if err != sql.ErrNoRows {
                http.NotFound(w, r)
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
    case "new":
        if r.Body == nil {
            http.Error(w, "Request body empty", http.StatusBadRequest)
            return
        }

        var p Post
        err := json.NewDecoder(r.Body).Decode(&p)

        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        err = createPost(p)

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

        var p Post
        err := json.NewDecoder(r.Body).Decode(&p)

        err = deletePost(path[2])

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

        var p Post 
        err := json.NewDecoder(r.Body).Decode(&p)

        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        err = editPost(path[2], p.content)

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


func loadPost(id string) (*Post, error) {
    var post Post

    query := `SELECT profileurl, authorurl, timestamp, content
            FROM post
            WHERE id = $1;`

    err := db.QueryRow(query, id).Scan(&(post.profileUrl),
            &(post.authorUrl), &(post.timestamp), &(post.content))

    if err != nil {
        return nil, err
    }

    return &post, nil
}

func createPost(post Post) error {
    query := `INSERT INTO post (profileurl, authorurl, timestamp, content)
            VALUES ($1, $2, $3, $4);`

    _, err := db.Exec(query, post.profileUrl, post.authorUrl,
            post.timestamp, post.content)

    return err
}

func deletePost(id string) error {
    query := `DELETE FROM post
            WHERE id = $1;`

    _, err := db.Exec(query, id)

    return err
}

func editPost(id string, content string) error {
    query := `UPDATE post
            SET content = $1
            WHERE id = $2;`

    _, err := db.Exec(query, content, id)

    return err
}
