package main

import (
	"database/sql"
	"fmt"
	"github.com/BurntSushi/toml"
	"html/template"
	"net/http"
	"os"

	_ "github.com/BurntSushi/toml"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
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

func init() {
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
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
	if err != nil {
		fmt.Println("Error creating users table:", err)
		os.Exit(1)
	} else {
		fmt.Println("Users table created successfully")
	}
}

func main() {
	defer func() {
		if db != nil {
			if err := db.Close(); err != nil {
				fmt.Println("Error closing database:", err)
			}
		}
	}()

	r := mux.NewRouter()

	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/register", registerHandler).Methods("GET", "POST")
	r.HandleFunc("/login", loginHandler).Methods("GET", "POST")
	r.HandleFunc("/logout", logoutHandler).Methods("GET", "POST")

	http.Handle("/", r)

	port := "8080"
	fmt.Printf("Starting server on :%s...\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println(err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"]
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	row := db.QueryRow("SELECT id, username FROM users WHERE id = ?", userID)
	err := row.Scan(&user.ID, &user.Username)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl, err := template.New("index").Parse(`
		<h1>Hello, {{.Username}}!</h1>
		<form action="/logout" method="post">
		  <button type="submit">Logout</button>
		</form>
		`)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, user)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		fmt.Println(db)

		// Проверка db на nil
		if db == nil {
			fmt.Println("Error: Database connection is nil")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, hashedPassword)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl, err := template.New("register").Parse(`
		<h1>Register</h1>
		<form action="/register" method="post">
		  <label for="username">Username:</label>
		  <input type="text" name="username" required>
		  <br>
		  <label for="password">Password:</label>
		  <input type="password" name="password" required>
		  <br>
		  <button type="submit">Register</button>
		</form>
		`)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		var user User
		row := db.QueryRow("SELECT id, username, password FROM users WHERE username = ?", username)
		err := row.Scan(&user.ID, &user.Username, &user.Password)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		session.Values["user_id"] = user.ID
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl, err := template.New("login").Parse(`
		<h1>Login</h1>
		<form action="/login" method="post">
		  <label for="username">Username:</label>
		  <input type="text" name="username" required>
		  <br>
		  <label for="password">Password:</label>
		  <input type="password" name="password" required>
		  <br>
		  <button type="submit">Login</button>
		</form>
		`)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	delete(session.Values, "user_id")
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
