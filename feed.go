package main

import(
    "net/http"
    "database/sql"
    _ "github.com/lib/pq"
    "encoding/json"
    "log"
    "time"
)

const PostsPerRequest int = 20

type FeedRequest struct {
    identifier  string
    mainFeed    bool
    before      time.Time
}

func feedHandler(w http.ResponseWriter, r *http.Request, url string) {

    if url != "" {
        http.NotFound(w, r)
        return
    }

    var req FeedRequest

    if r.Body == nil {
        http.Error(w, "Request body missing", http.StatusBadRequest)
        return
    }

    err := json.NewDecoder(r.Body).Decode(&req)

    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    var results map[int]Post

    if req.mainFeed {
        results, err = getFriendPosts(req.identifier, req.before)
    } else {
        results, err = getProfilePosts(req.identifier, req.before)
    }


    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    addHeaders(w, r)
    err = json.NewEncoder(w).Encode(results)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        log.Println(err)
    }

}

func getFriendPosts(userEmail string, before time.Time) (map[int]Post, error) {

    var (
        query string
        rows *sql.Rows
        err error
    )

    if before.IsZero() {
        query = `SELECT post.*
                FROM post, profile
                WHERE profile.email = $1 
                AND EXISTS (SELECT *
                            FROM connection
                            WHERE post.profileurl IN(connection.fromurl, connection.tourl)
                            AND post.authorurl IN(connection.fromurl, connection.tourl))
                ORDER BY post.timestamp DESC
                FETCH FIRST $2 ROWS ONLY;`
        rows, err = db.Query(query, userEmail, PostsPerRequest)
    } else {
        query = `SELECT post.*
                FROM post, profile
                WHERE profile.email = $1 
                AND post.timestamp < $2
                AND EXISTS (SELECT *
                            FROM connection
                            WHERE post.profileurl IN(connection.fromurl, connection.tourl)
                            AND post.authorurl IN(connection.fromurl, connection.tourl))
                ORDER BY post.timestamp DESC
                FETCH FIRST $3 ROWS ONLY;`
        rows, err = db.Query(query, userEmail, before, PostsPerRequest)
    }

    if err != nil {
        return nil, err
    }

    var results map[int]Post = make(map[int]Post)

    for rows.Next() {
        var post Post
        var id int
        err = rows.Scan(&id, &(post.profileUrl), &(post.authorUrl), &(post.timestamp), &(post.content))
        results[id] = post
    } 

    if err != nil {
        return nil, err
    }

    if len(results) < PostsPerRequest {
        if before.IsZero() {
            query = `SELECT post.*
                    FROM post, profile
                    WHERE profile.email = $1
                    AND EXISTS (SELECT * 
                                FROM connection
                                WHERE post.profileurl IN(connection.fromurl, connection.tourl))
                    ORDER BY post.timestamp DESC
                    FETCH FIRST $3 ROWS ONLY;`
            rows, err = db.Query(query, userEmail, PostsPerRequest - len(results))
        } else {
            query = `SELECT post.*
                    FROM post, profile
                    WHERE profile.email = $1
                    AND post.timestamp < $2
                    AND EXISTS (SELECT * 
                                FROM connection
                                WHERE post.profileurl IN(connection.fromurl, connection.tourl))
                    ORDER BY post.timestamp DESC
                    FETCH FIRST $2 ROWS ONLY;`
            rows, err = db.Query(query, before, userEmail, PostsPerRequest - len(results))
        }

        if err != nil {
            return nil, err
        }

        for rows.Next() {
            var post Post
            var id int
            err = rows.Scan(&id, &(post.profileUrl), &(post.authorUrl), &(post.timestamp), &(post.content))
            results[id] = post
        }

        if err != nil {
            return nil, err
        }
    }

    return results, nil
}

func getProfilePosts(profileUrl string, before time.Time) (map[int]Post, error) {

    var (
        query string
        rows *sql.Rows
        err error
    )

    if before.IsZero() {
        query = `SELECT *
                FROM post
                WHERE profileurl = $1 
                ORDER BY timestamp DESC
                FETCH FIRST $2 ROWS ONLY;`
        rows, err = db.Query(query, profileUrl)
    } else {
        query = `SELECT *
                FROM post
                WHERE profileurl = $1 
                AND timestamp < $2
                ORDER BY timestamp DESC
                FETCH FIRST $3 ROWS ONLY;`

        rows, err = db.Query(query, profileUrl, before)
    }
    if err != nil {
        return nil, err
    }

    var results map[int]Post = make(map[int]Post)

    for rows.Next() {
        var post Post
        var id int
        err = rows.Scan(&id, &(post.profileUrl), &(post.authorUrl), &(post.timestamp), &(post.content))
        results[id] = post
    }

    if err != nil {
        return nil, err
    }

    return results, nil
}
