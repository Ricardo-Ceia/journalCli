package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

var Users = map[int]User{
	1: {Name: "Alice", Password: "test123", Id: "1", JournalEntries: []string{"Today I learned Go.", "I love programming."}},
	2: {Name: "Bob", Password: "test123", Id: "2", JournalEntries: []string{"Go is great for web servers.", "I enjoy coding challenges."}},
}

type User struct {
	Name           string
	Id             string
	Password       string
	JournalEntries []string
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

	for _, user := range Users {
		if user.Name == username && user.Password == password {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(user.Id))
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

	newID := strconv.Itoa(len(Users) + 1)
	Users[(len(Users) + 1)] = User{Name: username, Password: password, Id: newID, JournalEntries: []string{}}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(newID))
}
