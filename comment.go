package main

import(
    "net/http"
    "database/sql"
    _ "github.com/lib/pq"
    "encoding/json"
    "log"
    "time"
)

type Comment struct {
    postId      int
    authorUrl   string
    timestamp   time.Time
    content     string
}

func commentHandler(w http.ResponseWriter, r *http.Request, url string) {

    path := connectPath.FindStringSubmatch(url)

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

        c, err := loadComment(path[2])

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

        err = json.NewEncoder(w).Encode(c)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            log.Println(err)
        }
    case "new":
        if r.Body == nil {
            http.Error(w, "Request body missing", http.StatusBadRequest)
            return
        }

        var c Comment
        err := json.NewDecoder(r.Body).Decode(&c)

        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        err = createComment(c)

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

        err := deleteComment(path[2])

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

        var c Comment
        err := json.NewDecoder(r.Body).Decode(&c)

        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        err = editComment(path[2], c.content)

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


func loadComment(id string) (*Comment, error) {

    query := `SELECT postid, authorurl, timestamp, content
            FROM comment
            WHERE id = $1;`

    var c Comment
    err := db.QueryRow(query, id).Scan(&(c.postId), &(c.authorUrl), 
            &(c.timestamp), &(c.content))

    if err != nil {
        return nil, err
    }

    return &c, nil
}

func createComment(comment Comment) error {
    query := `INSERT INTO post (postid, authorurl, timestamp, content)
            VALUES ($1, $2, $3, $4);`

    _, err := db.Exec(query, comment.postId, comment.authorUrl,
            comment.timestamp, comment.content)

    return err
}

func deleteComment(id string) error {
    query := `DELETE FROM comment
            WHERE id = $1;`

    _, err := db.Exec(query, id)

    return err
}

func editComment(id string, content string) error {
    query := `UPDATE comment
            SET content = $1
            WHERE id = $2;`

    _, err := db.Exec(query, content, id)

    return err
}
