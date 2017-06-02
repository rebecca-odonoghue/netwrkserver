package main

import(
    "net/http"
    "github.com/gorilla/mux"
    "database/sql"
    "bytes"
    _ "github.com/lib/pq"
    "encoding/json"
    "log"
    "strings"
)

const NumLiveResults int = 5
const NumResults int = 50

type Search struct {
    UserEmail   string  `json:"userEmail"`
    Live        bool    `json:"live"`
    Submit      bool    `json:"submit"`
}

type Result struct {
    URL         string  `json:"url"`
    FirstName   string  `json:"firstname"`
    LastName    string  `json:"lastname"`
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    query := vars["term"]
    var s Search

    if r.Body == nil {
        http.Error(w, "Request body missing", http.StatusBadRequest)
        return
    }

    err := json.NewDecoder(r.Body).Decode(&s)

    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if s.Submit {
        err = submitSearch(s.UserEmail, query)

        if err != nil {
            if err == sql.ErrNoRows {
                http.Error(w, err.Error(), http.StatusBadRequest)
            } else {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                log.Println(err)
            }
        }
    } else {
        var results []Result
        results, err = search(s, query)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        err = json.NewEncoder(w).Encode(results)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            log.Println(err)
        }
    }
}

func recentSearchHandler(w http.ResponseWriter, r *http.Request, query string) {

    if query != "" {
        http.NotFound(w, r)
        return
    }

    if r.Body == nil {
        http.Error(w, "Request body missing", http.StatusBadRequest)
        return
    }

    var email string
    err := json.NewDecoder(r.Body).Decode(&email)

    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    var results []Result
    results, err = getAllRecent(email)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    err = json.NewEncoder(w).Encode(results)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func saveSearchHandler(w http.ResponseWriter, r *http.Request, query string) {

    if query == "" {
        http.NotFound(w, r)
        return
    }

    if r.Body == nil {
        http.Error(w, "Request body missing", http.StatusBadRequest)
        return
    }

    var email string
    err := json.NewDecoder(r.Body).Decode(&email)

    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    err = submitSearch(email, query)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func search(s Search, searchString string) ([]Result, error) {

    if searchString == "" {
        return getAllRecent(s.UserEmail)
    }

    var (
        results []Result
        numResults int
        exp bytes.Buffer
        sep = ""
    )

    if s.Live {
        numResults = NumLiveResults
    } else {
        numResults = NumResults
    }

    exp.WriteString("%(")
    for _, word := range strings.Split(searchString, "+") {
        exp.WriteString(sep)
        exp.WriteString(strings.ToLower(word))
        sep = "|"
    }
    exp.WriteString(")%")

    r, err := searchRecent(s.UserEmail, exp.String(), numResults)

    if err != nil {
        return nil, err
    }

    for _,res := range r {
        results = append(results, res)
    }

    if len(r) < numResults {
        r, err = searchFriends(s.UserEmail, exp.String(), numResults - len(r))

        if (err != nil) {
            return nil, err
        }

        for _,res := range r {
            if !contains(results, res) {
                results = append(results, res)
            }
        }
    }

    if len(r) < numResults {
        r, err = searchAll(s.UserEmail, exp.String(), numResults - len(r))

        if err != nil {
            return nil, err
        }

        for _,res := range r {
            if !contains(results, res) {
                results = append(results, res)
            }
        }
    }



    return results, nil
}

func contains(list []Result, res Result) bool {

    for _, item := range list {
        if item.URL == res.URL {
            return true
        }
    }

    return false
}

func getAllRecent(userEmail string) ([]Result, error) {

    query := `SELECT profile.url, profile.firstname, profile.lastname
        FROM profile, search
        WHERE search.acctEmail = $1 
        AND profile.url = search.resultUrl
        ORDER BY search.timestamp DESC
        LIMIT $2;`

    rows, err := db.Query(query, userEmail, NumLiveResults)

    if err != nil {
        return nil, err
    }

    var results []Result

    for rows.Next() {
        var res Result
        err = rows.Scan(&(res.URL), &(res.FirstName), &(res.LastName))
        results = append(results, res)
    }

    if err != nil {
        return nil, err
    }

    return results, nil
}

func searchRecent(userEmail string, searchExp string, numResults int) ([]Result, error) {

    query := `SELECT profile.url, profile.firstname, profile.lastname
            FROM profile, search
            WHERE search.acctEmail = $1 
            AND profile.url = search.resultUrl
            AND (lower(profile.firstname) SIMILAR TO $2 
                OR lower(profile.lastname) SIMILAR TO $2
                OR lower(profile.email) = $3
                OR lower(profile.url) = $3)
            ORDER BY search.timestamp DESC
            LIMIT $4;`

    rows, err := db.Query(query, userEmail, searchExp, searchExp[2:len(searchExp)-2], numResults)

    if err != nil {
        return nil, err
    }

    var results []Result

    for rows.Next() {
        var res Result
        err = rows.Scan(&(res.URL), &(res.FirstName), &(res.LastName))
        results = append(results, res)
    }

    if err != nil {
        return nil, err
    }

    return results, nil
}

func searchFriends(userEmail string, searchExp string, numResults int) ([]Result, error) {

    query := `SELECT DISTINCT res.url, res.firstname, res.lastname
            FROM profile res, profile usr, connection
            WHERE usr.email = $1
            AND (lower(res.firstname) SIMILAR TO $2 
                OR lower(res.lastname) SIMILAR TO $2
                OR lower(res.email) = $3
                OR lower(res.url) = $3)
            AND (res.url IN(connection.fromurl, connection.tourl) 
            AND usr.url IN(connection.fromurl, connection.tourl))
            AND NOT usr.url = res.url
            LIMIT $4;`

    rows, err := db.Query(query, userEmail, searchExp,searchExp[2:len(searchExp)-2], numResults)

    if err != nil {
        return nil, err
    }

    var results []Result

    for rows.Next() {
        var res Result
        err = rows.Scan(&(res.URL), &(res.FirstName), &(res.LastName))
        results = append(results, res)
    }

    if err != nil {
        return nil, err
    }

    return results, nil
}

func searchAll(userEmail string, searchExp string, numResults int) ([]Result, error) {

    query := `SELECT profile.url, profile.firstname, profile.lastname
            FROM profile
            WHERE (lower(profile.firstname) SIMILAR TO $1 
                OR lower(profile.lastname) SIMILAR TO $1
                OR lower(profile.email) = $2
                OR lower(profile.url) = $2)
            LIMIT $3;`

    rows, err := db.Query(query, searchExp, searchExp[2:len(searchExp)-2], numResults)

    if err != nil {
        return nil, err
    }

    var results []Result

    for rows.Next() {
        var res Result
        err = rows.Scan(&(res.URL), &(res.FirstName), &(res.LastName))
        results = append(results, res)
    }

    if err != nil {
        return nil, err
    }

    return results, nil
}

func submitSearch(userEmail string, result string) error {
    query := `INSERT INTO search (acctEmail, resultUrl)
            VALUES ($1, $2);`

    _, err := db.Exec(query, userEmail, result)

    return err
}
