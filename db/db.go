package db

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq"
)

var (
	instance *sql.DB
	once     sync.Once
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "password"
	dbname   = "journaldb"
)

func InitDB() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	instance = db
	fmt.Println("Connected to database ✅")
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
			instance = InitDB()
		})
	}
	return instance
}
