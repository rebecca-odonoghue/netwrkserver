package main

import (
    "net/http"
    _ "github.com/lib/pq"
    "time"
    "golang.org/x/crypto/bcrypt"
    "encoding/json"
)

type Account struct {
    email       string
    dob         time.Time
}

type Registration struct {
    account     Account
    password    string
}

func authenticationHandler(w http.ResponseWriter, r *http.Request, url string) {

    if url != "" {
        http.NotFound(w, r)
        return
    }

    ok := checkAuthorisation(w,r)

    if ok {
        w.WriteHeader(200)
    }

}
func registrationHandler(w http.ResponseWriter, r *http.Request, url string) {

    var reg Registration

    if r.Body == nil {
        http.Error(w, "Request body missing", http.StatusBadRequest)
        return
    }

    err := json.NewDecoder(r.Body).Decode(&reg)

    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    err = createAccount(reg.account.email, reg.account.dob, reg.password)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

func accountHandler(w http.ResponseWriter, r *http.Request, url string) {

    switch url {
    case "delete":

        var email string

        if r.Body == nil {
            http.Error(w, "Request body missing", http.StatusBadRequest)
            return
        }

        err := json.NewDecoder(r.Body).Decode(&email)

        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        err = deleteAccount(email)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    case "modify":

        var reg Registration

        if r.Body == nil {
            http.Error(w, "Request body missing", http.StatusBadRequest)
            return
        }

        err := json.NewDecoder(r.Body).Decode(&reg)

        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        err = changePassword(reg.account.email, reg.password)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    default:
        http.NotFound(w, r)
    }
}

func createAccount(email string, dob time.Time, password string) error {
    query := `INSERT INTO account (email, dob, password)
            VALUES ($1, $2, $3);`

    hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

    if err == nil {
        _, err = db.Exec(query, email, dob, hashedPwd)
    }

    return err
}

func changePassword(email string, password string) error {

    query := `UPDATE account
            SET password = $1
            WHERE email = $2;`

    hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

    if err == nil {
        _, err = db.Exec(query, hashedPwd, email)
    }

    return err
}

func deleteAccount(email string) error {
    query := `DELETE FROM account
            WHERE email = $1;`

    _, err := db.Exec(query, email)

    return err
}

func authenticate(email string, password string) error {
    query := `SELECT password
            FROM account
            WHERE email = $1;`

    var hashedPwd []byte

    err := db.QueryRow(query, email).Scan(&hashedPwd)

    if err == nil {
        err = bcrypt.CompareHashAndPassword(hashedPwd, []byte(password))
    }

    return err
}
