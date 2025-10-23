package main

import (
	"fmt"
	"journalCli/db"
	"journalCli/handlers"
	"net/http"
)

func server() {
	dbPath := "../journal.db"

	database := db.InitDB(dbPath)

	defer db.CloseDB(database)

	http.HandleFunc("/signup", handlers.SignUpHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	fmt.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func main() {
	server()
}
