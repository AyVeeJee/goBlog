package models

import "time"

type User struct {
	ID       int
	Username string
	Password string
}

type Post struct {
	ID   int       `json:"id"`
	Name string    `json:"name"`
	Text string    `json:"text"`
	Date time.Time `json:"date"`
}
