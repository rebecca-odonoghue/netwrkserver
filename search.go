package main

import(
    "net/http"
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
    userEmail   string
    live        bool
    submit      bool
}

type Result struct {
    url         string
    firstname   string
    lastname    string
}

func searchHandler(w http.ResponseWriter, r *http.Request, query string) {

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

    if s.submit {
        err = submitSearch(s.userEmail, query)

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
        return getAllRecent(s.userEmail)
    }

    var (
        results []Result
        numResults int
        exp bytes.Buffer
        sep = ""
    )

    if s.live {
        numResults = NumLiveResults
    } else {
        numResults = NumResults
    }

    exp.WriteString("%(")
    for _, word := range strings.Split(searchString, "+") {
        exp.WriteString(sep)
        exp.WriteString(word)
        sep = "|"
    }
    exp.WriteString(")%")

    r, err := searchRecent(s.userEmail, exp.String(), numResults)

    if err != nil {
        return nil, err
    }

    for _,res := range r {
        results = append(results, res)
    }

    if len(r) < numResults {
        r, err = searchFriends(s.userEmail, exp.String(), numResults - len(r))

        if (err != nil) {
            return nil, err
        }
    }

    if len(r) < numResults {
        r, err = searchAll(s.userEmail, exp.String(), numResults - len(r))

        if err != nil {
            return nil, err
        }
    }

    return results, nil
}

func getAllRecent(userEmail string) ([]Result, error) {

    query := `SELECT profile.url, profile.firstname, profile.lastname
        FROM profile, search
        WHERE search.acctEmail = $1 
        AND profile.url = search.resultUrl
        ORDER BY search.timestamp DESC
        FETCH FIRST $2 ROWS ONLY;`

    rows, err := db.Query(query, userEmail, NumLiveResults)

    if err != nil {
        return nil, err
    }

    var results []Result

    for rows.Next() {
        var res Result
        err = rows.Scan(&(res.url), &(res.firstname), &(res.lastname))
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
            AND (profile.firstname SIMILAR TO $2 
                OR profile.lastname SIMILAR TO $2
                OR profile.email SIMILAR TO $2
                OR profile.url SIMILAR TO $2)
            ORDER BY search.timestamp DESC
            FETCH FIRST $3 ROWS ONLY;`

    rows, err := db.Query(query, userEmail, searchExp, numResults)

    if err != nil {
        return nil, err
    }

    var results []Result

    for rows.Next() {
        var res Result
        err = rows.Scan(&(res.url), &(res.firstname), &(res.lastname))
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
            AND (res.firstname SIMILAR TO $2 
                OR res.lastname SIMILAR TO $2
                OR res.email SIMILAR TO $2
                OR res.url SIMILAR TO $2)
            AND (res.url IN(connection.fromurl, connection.tourl) 
            AND usr.url IN(connection.fromurl, connection.tourl))
            AND NOT usr.url = res.url
            ORDER BY search.timestamp DESC
            FETCH FIRST $3 ROWS ONLY;`

    rows, err := db.Query(query, userEmail, searchExp, numResults)

    if err != nil {
        return nil, err
    }

    var results []Result

    for rows.Next() {
        var res Result
        err = rows.Scan(&(res.url), &(res.firstname), &(res.lastname))
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
            WHERE (profile.firstname SIMILAR TO $1 
                OR profile.lastname SIMILAR TO $1
                OR profile.email SIMILAR TO $1
                OR profile.url SIMILAR TO $1)
            ORDER BY search.timestamp DESC
            FETCH FIRST $2 ROWS ONLY;`

    rows, err := db.Query(query, searchExp, numResults)

    if err != nil {
        return nil, err
    }

    var results []Result

    for rows.Next() {
        var res Result
        err = rows.Scan(&(res.url), &(res.firstname), &(res.lastname))
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
