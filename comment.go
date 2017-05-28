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

func commentHandler(w http.ResponseWriter, r *http.Request, id string) {
    log.Println("Comment " + id + "requested")

    comment, err := loadComment(id)

    if err != nil {
        if err != sql.ErrNoRows {
            http.Error(w, err.Error(), http.StatusBadRequest)
            log.Println(err)
            return
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            log.Println(err)
            return
        }
    }

    addHeaders(w, r)
    err = json.NewEncoder(w).Encode(comment)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        log.Println(err)
    }

}


func loadComment(id string) (*Comment, error) {
    var (
        postId      int
        authorUrl   string
        timestamp   time.Time
        content     string
    )

    query := `SELECT profile-url, author-url, timestamp, content
            FROM comment
            WHERE id = $1;`

    err := db.QueryRow(query, id).Scan(&postId, &authorUrl, &timestamp, &content)

    if err != nil {
        return nil, err
    }

    return &Comment {postId: postId,
            authorUrl: authorUrl,
            timestamp: timestamp,
            content: content}, nil
}

func createComment(id int, comment *Comment) error {
    query := `INSERT INTO post (post-id, author-url, timestamp, content)
            VALUES ($1, $2, $3, $4);`

    _, err := db.Exec(query, comment.postId, comment.authorUrl,
            comment.timestamp, comment.content)

    return err
}

func deleteComment(id int) error {
    query := `DELETE FROM comment
            WHERE id = $1;`

    _, err := db.Exec(query, id)

    return err
}

func editComment(id int, content string) error {
    query := `UPDATE comment
            SET content = $1
            WHERE id = $2;`

    _, err := db.Exec(query, content, id)

    return err
}
