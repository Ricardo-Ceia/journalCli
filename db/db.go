package db

import (
	"database/sql"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	instance *sql.DB
	once     sync.Once
)

func InitDB(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
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

func GetDB() *sql.DB {
	if instance == nil {
		//if the databse is not initialized, initialize it (once.Do ensures that this code only runs once)
		once.Do(func() {
			instance = InitDB("db.sqlite3")
		})
	}
	return instance
}
