package api

import (
	"database/sql"
	"encoding/json"
	"goBlog/src/common/models"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type BlogController struct {
	db *sql.DB
}

func NewBlogController(db *sql.DB) *BlogController {
	return &BlogController{
		db: db,
	}
}

func (bc *BlogController) CreatePost(w http.ResponseWriter, r *http.Request) {
	var newPost models.Post
	err := json.NewDecoder(r.Body).Decode(&newPost)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	result, err := bc.db.Exec("INSERT INTO posts (name, text, date) VALUES (?, ?, ?)", newPost.Name, newPost.Text, time.Now())
	if err != nil {
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}

	newPostID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Error retrieving post ID", http.StatusInternalServerError)
		return
	}
	newPost.ID = int(newPostID)
	newPost.Date = time.Now()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPost)
}

func (bc *BlogController) GetPosts(w http.ResponseWriter, r *http.Request) {
	rows, err := bc.db.Query("SELECT id, name, text, date FROM posts")
	if err != nil {
		http.Error(w, "Error retrieving posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.ID, &post.Name, &post.Text, &post.Date)
		if err != nil {
			http.Error(w, "Error scanning posts", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (bc *BlogController) GetPostByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	postID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var post models.Post
	err = bc.db.QueryRow("SELECT id, name, text, date FROM posts WHERE id = ?", postID).Scan(&post.ID, &post.Name, &post.Text, &post.Date)
	if err == sql.ErrNoRows {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error retrieving post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func (bc *BlogController) UpdatePost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	postID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var updatedPost models.Post
	err = json.NewDecoder(r.Body).Decode(&updatedPost)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	_, err = bc.db.Exec("UPDATE posts SET name=?, text=?, date=? WHERE id=?", updatedPost.Name, updatedPost.Text, time.Now(), postID)
	if err != nil {
		http.Error(w, "Error updating post", http.StatusInternalServerError)
		return
	}

	updatedPost.ID = postID
	updatedPost.Date = time.Now()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedPost)
}

func (bc *BlogController) DeletePost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	postID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	_, err = bc.db.Exec("DELETE FROM posts WHERE id=?", postID)
	if err != nil {
		http.Error(w, "Error deleting post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
