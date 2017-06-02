package main

import (
    "net/http"
    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
    "time"
    "golang.org/x/crypto/bcrypt"
    "encoding/json"
    "log"
)

type Account struct {
    Email       string      `json:"email"`
    DOB         time.Time   `json:"dob"`
}

type Registration struct {
    Account     Account     `json:"account"`
    Password    string      `json:"password"`
}

func authenticationHandler(w http.ResponseWriter, r *http.Request, url string) {

    if url != "" {
        http.NotFound(w, r)
        return
    }

    email, ok := checkAuthorisation(w,r)

    if ok {

        query := `SELECT firstname, lastname, url
                FROM profile
                WHERE email = $1;`

        var firstname string
        var lastname string
        var path string
        row := db.QueryRow(query, email)
        err := row.Scan(&firstname, &lastname, &path)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }

        err = json.NewEncoder(w).Encode(struct {
            FirstName string `json:"firstname"`
            LastName string `json:"lastname"`
            URL string `json:"url"`
        }{
            FirstName: firstname,
            LastName: lastname,
            URL: path,
        })

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }

        log.Println("Authentication successful")
    }

}
func registrationHandler(w http.ResponseWriter, r *http.Request, url string) {

    log.Println("Registration requested");
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
    log.Println("Request processed")
    err = createAccount(reg.Account.Email, reg.Account.DOB, reg.Password)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    log.Println("Account created");
}

func accountHandler(w http.ResponseWriter, r *http.Request) {
    
    vars := mux.Vars(r)
    action := vars["action"]

    switch action {
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

        email, ok := checkAuthorisation(w,r)
        
        if !ok {
            return
        }

        var acct struct { 
            Pwd string `json:"password"`
        }

        if r.Body == nil {
            http.Error(w, "Request body missing", http.StatusBadRequest)
            return
        }

        err := json.NewDecoder(r.Body).Decode(&acct)

        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        err = changePassword(email, acct.Pwd)

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
