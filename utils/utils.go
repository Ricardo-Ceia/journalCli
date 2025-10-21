package utils

import (
	"fmt"
	"log"
	"os"
)

func ValidateCredentials(username, password, confirmPassword string) (bool, error) {
	if password != confirmPassword {
		return false, fmt.Errorf("Passwords do not match")
	}

	if len("username") < 3 || len("username") > 20 {
		return false, fmt.Errorf("Username too small or too big please choose a username with more than 3 characters and less than 20")
	}

	if len("password") < 6 || len("password") > 50 {
		return false, fmt.Errorf("Password too small or too big please choose a password with more than 6 characters and less than 50")
	}

	return true, nil
}

var DebugLog *log.Logger

func InitDebugFile() {
	f, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Error creating debug file:", err)
		return
	}
	defer f.Close()

	DebugLog = log.New(f, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
}
