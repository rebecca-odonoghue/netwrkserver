package main

import (
    "net/http"
    _ "github.com/lib/pq"
    "time"
    "golang.org/x/crypto/bcrypt"
)

type Account struct {
    email       string
    dob         time.Time
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

func createAccount(email string, dob time.Time, password string) error {
    query := `INSERT INTO account (email, dob, password)
            VALUES ($1, $2, $3);`

    hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

    if err == nil {
        _, err = db.Exec(query, email, dob, hashedPwd)
    }

    return err
}

func deleteAccount(email string) error {
    query := `DELETE FROM account
            WHERE email = $1`

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
