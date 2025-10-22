package db

import (
	"database/sql"
	"fmt"
	"journalCli/handlers"
	"log"
)

func InitDB() *sql.DB {
	db, err := sql.Open("sqlite3", "database=journal.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func CloseDB(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
}

func GetUserByID(id string) (handlers.User, error) {
	for _, user := range handlers.Users {
		if user.Id == id {
			return user, nil
		}
	}
	return handlers.User{}, fmt.Errorf("User not found")
}
