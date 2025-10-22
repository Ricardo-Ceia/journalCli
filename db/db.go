package db

import (
	"database/sql"
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
