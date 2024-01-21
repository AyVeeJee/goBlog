package main

import (
	"database/sql"
	"fmt"
	"goBlog/src/api"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
)

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

func main() {
	var config Config
	if _, err := toml.DecodeFile("../../src/config/config.toml", &config); err != nil {
		fmt.Println("Error reading config file:", err)
		os.Exit(1)
	}

	//store := sessions.NewCookieStore([]byte(config.Session.Key))
	db, err := sql.Open(config.Database.Driver, config.Database.ConnectionString)
	if err != nil {
		fmt.Println("Error opening database:", err)
		os.Exit(1)
	}

	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT,
        password TEXT
    )`)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT,
        text TEXT,
		date DATETIME
    )`)

	authController := api.NewAuthController(db)
	blogController := api.NewBlogController(db)

	r := chi.NewRouter()

	r.Route("/{user_id}", func(r chi.Router) {
		r.Get("/", authController.HomeHandler)
	})

	r.Route("/api/register", func(r chi.Router) {
		r.Post("/", authController.RegisterHandler)
	})

	r.Route("/api/login", func(r chi.Router) {
		r.Post("/", authController.LoginHandler)
	})

	r.Route("/api/logout", func(r chi.Router) {
		r.Post("/", authController.LogoutHandler)
	})

	r.Route("/api/post", func(r chi.Router) {
		r.Post("/", blogController.CreatePost)
		r.Get("/", blogController.GetPosts)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", blogController.GetPostByID)
			r.Put("/", blogController.UpdatePost)
			r.Delete("/", blogController.DeletePost)
		})
	})

	fmt.Println("Starting server on :8080")
	http.ListenAndServe(":8080", r)
}
