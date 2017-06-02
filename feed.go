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
    Identifier  string      `json:"identifier"`
    MainFeed    bool        `json:"mainFeed"`
    Before      time.Time   `json:"before"`
}

func feedHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("feedHandler");

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

    var results []Post

    if req.MainFeed {
        results, err = getFriendPosts(req.Identifier, req.Before)
    } else {
        results, err = getProfilePosts(req.Identifier, req.Before)
    }


    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    err = json.NewEncoder(w).Encode(results)
    log.Println(len(results))

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        log.Println(err)
    }
    log.Println("Feed sent")
    //log.Println(results));
}

func getFriendPosts(userEmail string, before time.Time) ([]Post, error) {
    log.Println("getFriendPosts");
    log.Println(before);
    var (
        query string
        rows *sql.Rows
        err error
    )

    if before.IsZero() {
        query = `SELECT p.id, p.profileurl, p.authorurl, p.timestamp, p.content
                FROM post p, profile q
                WHERE q.email = '$1' 
                AND EXISTS (SELECT *
                            FROM connection c
                            WHERE q.url IN(c.fromurl, c.tourl)
                            AND (p.profileurl IN(c.fromurl, c.tourl)
                                OR p.authorurl IN(c.fromurl, c.tourl)))
                ORDER BY p.timestamp DESC
                LIMIT $2;`
        rows, err = db.Query(query, userEmail, PostsPerRequest)
    } else {
        query = `SELECT p.id, p.profileurl, p.authorurl, p.timestamp, p.content
                FROM post p, profile q
                WHERE q.email = $1 
                AND p.timestamp < $2
                AND EXISTS (SELECT *
                            FROM connection c
                            WHERE q.url IN(c.fromurl, c.tourl)
                            AND (p.profileurl IN(c.fromurl, c.tourl)
                                OR p.authorurl IN(c.fromurl, c.tourl)))
                ORDER BY p.timestamp DESC
                LIMIT $3;`
        rows, err = db.Query(query, userEmail, before, PostsPerRequest)
    }

    if err != nil {
        return nil, err
    }

    var results []Post

    for rows.Next() {
        var post Post
        err = rows.Scan(&post.ID, &post.ProfileUrl, &post.AuthorUrl, &post.Timestamp, &post.Content)
        if err != nil {
            return nil, err
        }

        if !containsPost(results, post) {
            results = append(results, post)
        }
    } 

    if err != nil {
        return nil, err
    }
/*
    if len(results) < PostsPerRequest {
        if before.IsZero() {
            query = `SELECT p.id, p.profileurl, p.authorurl, p.timestamp, p.content
                    FROM post p, profile q
                    WHERE q.email = $1
                    AND EXISTS (SELECT * 
                                FROM connection c
                                WHERE p.profileurl IN(c.fromurl, c.tourl))
                    ORDER BY p.timestamp DESC
                    LIMIT $2;`
            rows, err = db.Query(query, userEmail, PostsPerRequest - len(results))
        } else {
            query = `SELECT p.id, p.profileurl, p.authorurl, p.timestamp, p.content
                    FROM post p, profile q
                    WHERE q.email = $1
                    AND p.timestamp < $2
                    AND EXISTS (SELECT * 
                                FROM connection c
                                WHERE p.profileurl IN(c.fromurl, c.tourl))
                    ORDER BY p.timestamp DESC
                    LIMIT $3;`
            rows, err = db.Query(query, userEmail, before, PostsPerRequest - len(results))
        }

        if err != nil {
            return nil, err
        }

        for rows.Next() {
            var post Post
            err = rows.Scan(&post.ID, &post.ProfileUrl, &post.AuthorUrl, &post.Timestamp, &post.Content)
            if err != nil {
                return nil, err
            }

            if !containsPost(results, post) {
                results = append(results, post)
            }
        }

        if err != nil {
            return nil, err
        }
    }
*/
    return results, nil
}

func containsPost(list []Post, post Post) bool {

    for _, item := range list {
        if item.ID == post.ID {
            return true
        }
    }

    return false
}

func getProfilePosts(profileUrl string, before time.Time) ([]Post, error) {
    var (
        query string
        rows *sql.Rows
        err error
    )

    if before.IsZero() {
        query = `SELECT id, profileurl, authorurl, timestamp, content
                FROM post
                WHERE profileurl = $1 
                ORDER BY timestamp DESC
                LIMIT $2;`
        rows, err = db.Query(query, profileUrl, PostsPerRequest)
    } else {
        query = `SELECT id, profileurl, authorurl, timestamp, content
                FROM post
                WHERE profileurl = $1 
                AND timestamp < $2
                ORDER BY timestamp DESC
                LIMIT $3;`

        rows, err = db.Query(query, profileUrl, before, PostsPerRequest)
    }
    if err != nil {
        return nil, err
    }

    var results []Post

    for rows.Next() {
        var post Post
        err = rows.Scan(&post.ID, &(post.ProfileUrl), &(post.AuthorUrl), &(post.Timestamp), &(post.Content))
        results = append(results, post)
    }

    if err != nil {
        return nil, err
    }

    return results, nil
}
