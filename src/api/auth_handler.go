package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"goBlog/src/common/models"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type AuthController struct {
	db *sql.DB
}

func NewAuthController(db *sql.DB) *AuthController {
	return &AuthController{
		db: db,
	}
}

type registerUser struct {
	Username string
	Password string
}

func (ctrl *AuthController) HomeHandler(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if userID != "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var user models.User
	row := ctrl.db.QueryRow("SELECT id, username FROM users WHERE id = ?", userID)
	err := row.Scan(&user.ID, &user.Username)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	jsonResponse := map[string]interface{}{
		"message": "Hello, " + user.Username + "!",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonResponse)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method == "POST" {
		username := r.PostFormValue("username")
		password := r.PostFormValue("password")

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		_, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, hashedPassword)
		if err != nil {
			//
		}

		jsonResponse := map[string]interface{}{
			"message": "Registration successful",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jsonResponse)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore, db *sql.DB, user *models.User) {
	session, _ := store.Get(r, "session")
	if r.Method == "POST" {
		username := r.PostFormValue("username")
		password := r.PostFormValue("password")

		var loginUser models.User
		row := db.QueryRow("SELECT id, username, password FROM users WHERE username = ?", username)
		err := row.Scan(&loginUser.ID, &loginUser.Username, &loginUser.Password)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(loginUser.Password), []byte(password))
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		session.Values["user_id"] = loginUser.ID
		session.Save(r, w)

		jsonResponse := map[string]interface{}{
			"message": "Login successful",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jsonResponse)

		return
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore) {
	session, _ := store.Get(r, "session")
	delete(session.Values, "user_id")
	session.Save(r, w)

	jsonResponse := map[string]interface{}{
		"message": "Logout successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonResponse)
}
