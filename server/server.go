package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type user struct {
	name           string
	id             int
	password       string
	journalEntries []string
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

//TODO: Implement the login response structure
/*
type LoginRespose struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token"`
}
*/
var users = map[int]user{
	1: {name: "Alice", password: "test123", id: 1, journalEntries: []string{"Today I learned Go.", "I love programming."}},
	2: {name: "Bob", password: "test123", id: 2, journalEntries: []string{"Go is great for web servers.", "I enjoy coding challenges."}},
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginReq LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginReq)

	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	username := loginReq.Username
	password := loginReq.Password

	for _, user := range users {
		if user.name == username && user.password == password {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(strconv.Itoa(user.id)))
			return
		}
	}
	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
}

func server() {
	http.HandleFunc("/auth", authHandler)
	fmt.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func main() {
	server()
}
