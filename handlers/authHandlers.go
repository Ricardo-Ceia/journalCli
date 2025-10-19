package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

var users = map[int]user{
	1: {name: "Alice", password: "test123", id: 1, journalEntries: []string{"Today I learned Go.", "I love programming."}},
	2: {name: "Bob", password: "test123", id: 2, journalEntries: []string{"Go is great for web servers.", "I enjoy coding challenges."}},
}

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

type SignupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
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

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var signupReq SignupRequest

	err := json.NewDecoder(r.Body).Decode(&signupReq)

	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	username := signupReq.Username
	password := signupReq.Password

	newID := len(users) + 1
	users[newID] = user{name: username, password: password, id: newID, journalEntries: []string{}}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(strconv.Itoa(newID)))
}
