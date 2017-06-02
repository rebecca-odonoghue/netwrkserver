package main

import (
    "net/http"
    "github.com/gorilla/mux"
    "database/sql"
    _ "github.com/lib/pq"
    "log"
    "encoding/json"
    "time"
)

type Post struct {
    ProfileUrl  string      `json:"profileUrl"`
    AuthorUrl   string      `json:"authorUrl"`
    ID          int         `json:"id"`
    Timestamp   time.Time   `json:"timestamp"`
    Content     string      `json:"content"`
}

func postHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("postHandler")
    vars := mux.Vars(r)
    action := vars["action"]
    id := vars["id"]

    switch action {
    case "get":
        if id == "" {
            http.NotFound(w, r)
            return
        }

        p, err := loadPost(id)

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
        if id == "" {
            http.NotFound(w, r)
            return
        }

        //var p Post
        //err := json.NewDecoder(r.Body).Decode(&p)

        err := deletePost(id)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
    case "modify":
        if r.Body == nil || id == "" {
            http.Error(w, "Request incomplete", http.StatusBadRequest)
            return
        }

        var p Post 
        err := json.NewDecoder(r.Body).Decode(&p)

        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        err = editPost(id, p.Content)

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

    err := db.QueryRow(query, id).Scan(&(post.ProfileUrl),
            &(post.AuthorUrl), &(post.Timestamp), &(post.Content))

    if err != nil {
        return nil, err
    }

    return &post, nil
}

func createPost(post Post) error {
    query := `INSERT INTO post (profileurl, authorurl, content)
            VALUES ($1, $2, $3);`

    _, err := db.Exec(query, post.ProfileUrl, post.AuthorUrl, post.Content)

    return err
}

func deletePost(id string) error {
    log.Println("delete: "+id)
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
