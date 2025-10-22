package db

import (
	"database/sql"
	"strconv"
)

type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Username      string `json:"username"`
	Password_hash string `json:"password_hash"`
}

func CreateUser(db *sql.DB, username, email, password_hash string) (string, error) {
	res, err := db.Exec(`Insert Into users (username,email,password_hash)`, username, email, password_hash)
	if err != nil {
		return "", err
	}

	intID, err := res.LastInsertId()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(intID, 10), nil
}

func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	var user User
	row := db.QueryRow(`SELECT id, email, username FROM users WHERE email = ?`, email)
	if err := row.Scan(&user.ID, &user.Email, &user.Username); err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByID(db *sql.DB, id string) (*User, error) {
	var user User
	row := db.QueryRow(`SELECT username, email FROM users WHERE id = ?`, id)
	if err := row.Scan(&user.Username, &user.Email); err != nil {
		return nil, err
	}
	return &user, nil
}
