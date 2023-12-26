package main

import (
	"database/sql"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"os"
)

type User struct {
	ID       int
	Username string
	Password string
}

type Config struct {
	App      AppConfig
	Database DBConfig
	Session  SessionConfig
}

type AppConfig struct {
	Port int
}

type DBConfig struct {
	Driver           string
	ConnectionString string
}

type SessionConfig struct {
	Key string
}

var db *sql.DB
var config Config
var user User
var store *sessions.CookieStore
var err error

func main() {
	if _, err := toml.DecodeFile("/config/config.toml", &config); err != nil {
		fmt.Println("Error reading config file:", err)
		os.Exit(1)
	}

	store = sessions.NewCookieStore([]byte(config.Session.Key))
	db, err = sql.Open(config.Database.Driver, config.Database.ConnectionString)

	if err != nil {
		fmt.Println("Error opening database:", err)
		os.Exit(1)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT,
        password TEXT
    )`)

	defer func() {
		if db != nil {
			if err := db.Close(); err != nil {
				fmt.Println("Error closing database:", err)
			}
		}
	}()

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		HomeHandler(w, r, store, db, &user)
	}).Methods("GET")

	r.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		RegisterHandler(w, r, db)
	}).Methods("POST")

	r.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		LoginHandler(w, r, store, db, &user)
	}).Methods("POST")

	r.HandleFunc("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		LogoutHandler(w, r, store)
	}).Methods("POST")

	http.Handle("/", r)

	port := "8080"
	fmt.Printf("Starting server on :%s...\n", port)
	http.ListenAndServe(":"+port, nil)
}
