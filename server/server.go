package main

import (
	"fmt"
	"net/http"
	"strconv"
)

type user struct {
	name           string
	id             int
	journalEntries []string
}

var users = map[int]user{
	1: {name: "Alice", id: 1, journalEntries: []string{"Today I learned Go.", "I love programming."}},
	2: {name: "Bob", id: 2, journalEntries: []string{"Go is great for web servers.", "I enjoy coding challenges."}},
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userId, err := strconv.Atoi(r.URL.Query().Get("userId"))
	fmt.Println("userId:", userId)
	if err != nil {
		http.Error(w, "Invalid userId parameter", http.StatusBadRequest)
		return
	}

	if userId == 0 {
		http.Error(w, "Missing userId parameter", http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, users[userId].journalEntries)
}

func server() {
	http.HandleFunc("/user", handler)
	fmt.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func main() {
	server()
}
