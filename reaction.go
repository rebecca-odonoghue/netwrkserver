package main

import(
    "net/http"
    _ "github.com/lib/pq"
    "encoding/json"
    "log"
    "regexp"
)

type Reaction struct {
    identifier  int
    authorUrl   string
    toPost      bool
    isLike      bool
}

var reactPath = regexp.MustCompile("^(new|delete|modify)$")

func reactionHandler(w http.ResponseWriter, r *http.Request, url string) {
    
    path := validPath.FindStringSubmatch(url)

    if  path == nil {
        http.NotFound(w, r)
        return
    }

    if r.Body == nil {
        http.Error(w, "Request body missing", http.StatusBadRequest)
        return
    }

    var react Reaction

    err := json.NewDecoder(r.Body).Decode(&react)

    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    switch path[1] {
    case "new":
        err = newReaction(react.identifier, react.authorUrl, react.toPost, react.isLike)
    case "delete":
        err = deleteReaction(react.identifier, react.authorUrl, react.toPost)
    case "modify":
        err = modifyReaction(react.identifier, react.authorUrl, react.toPost, react.isLike)
    default: 
        http.NotFound(w, r)
        return
    }

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        log.Println(err)
        return
    }

    w.WriteHeader(200)
}


func getReactions(identifier int, toPost bool) ([]Reaction, error) {

    var query string

    if toPost {
        query = `SELECT * 
                FROM reaction
                WHERE post = true
                AND postid = $1;`
    } else {
        query = `SELECT * 
                FROM reaction
                WHERE post = false
                AND commentid = $1;`
    }

    rows, err := db.Query(query, identifier)

    if err != nil {
        return nil, err
    }

    var reactions []Reaction

    for rows.Next() {
        var r Reaction
        err = rows.Scan(&(r.identifier), &(r.authorUrl), &(r.toPost), &(r.isLike))
        reactions = append(reactions, r)
    }

    if err != nil {
        return nil, err
    }

    return reactions, nil
}

func newReaction(identifier int, userUrl string, toPost bool, isLike bool) error {
    
    var query string
    
    if toPost {
        query = `INSERT INTO reaction (authorurl, like, post, postid, commentid)
            VALUES ($1, $2, true, $3, 0);`
    } else {
        query = `INSERT INTO reaction (authorurl, like, post, postid, commentid)
            VALUES ($1, $2, false, 0, $3);`
    }

    _, err := db.Exec(query, userUrl, isLike, identifier)

    return err
}

func deleteReaction(identifier int, userUrl string, toPost bool) error {
    var query string
    
    if toPost {
        query = `DELETE FROM reaction 
                WHERE authorurl = $1
                AND post = true
                AND postid = $2;`
    } else {
        query = `DELETE FROM reaction 
                WHERE authorurl = $1
                AND post = false
                AND commentid = $2;`
    }

    _, err := db.Exec(query, userUrl, identifier)

    return err
}

func modifyReaction(identifier int, userUrl string, toPost bool, isLike bool) error {

    var query string

    if toPost {
        query = `UPDATE reaction
                SET like = $1
                WHERE authorurl = $2
                AND postid = $3;`
    } else {
        query = `UPDATE reaction
                SET like = $1
                WHERE authorurl = $2
                AND commentid = $3;`
    }

    _, err := db.Exec(query, isLike, userUrl, identifier)

    return err
}
