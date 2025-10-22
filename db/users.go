package db

import (
	"database/sql"
)

type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Username      string `json:"username"`
	Password_hash string `json:"password_hash"`
}

func CreateUser(db *sql.DB, username, email, password_hash string) (int64, error) {
	res, err := db.Exec(`Insert Into users (username,email,password_hash)`, username, email, password_hash)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	var user User
	row := db.QueryRow(`SELECT id, email, username FROM users WHERE email = ?`, email)
	if err := row.Scan(&user.ID, &user.Email, &user.Username); err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByID(db *sql.DB, id int64) (string, string, error) {
	row := db.QueryRow(`SELECT username, email FROM users WHERE id = ?`, id)
	var username, email string
	if err := row.Scan(&username, &email); err != nil {
		return "", "", err
	}
	return username, email, nil
}
