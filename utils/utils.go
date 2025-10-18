package utils

import (
	"fmt"
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
