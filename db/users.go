package db

import (
	"database/sql"
	"fmt"
	"strconv"
)

type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Username      string `json:"username"`
	Password_hash string `json:"password_hash"`
}

func CreateUser(db *sql.DB, username, email, password_hash string) (*User, error) {
	var id int
	query := `INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id`
	err := db.QueryRow(query, username, email, password_hash).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	return &User{
		ID:            strconv.Itoa(id),
		Email:         email,
		Username:      username,
		Password_hash: "",
	}, nil
}

func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	var user User
	query := `SELECT id, email, username, password_hash FROM users WHERE email = $1`
	err := db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Username, &user.Password_hash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByID(db *sql.DB, id string) (*User, error) {
	var user User
	query := `SELECT username, email FROM users WHERE id = $1`
	err := db.QueryRow(query, id).Scan(&user.Username, &user.Email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
