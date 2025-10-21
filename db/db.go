package db

import (
	"fmt"
	"journalCli/handlers"
)

var users = handlers.Users

func GetUserByID(id string) (handlers.User, error) {
	for _, user := range users {
		if user.Id == id {
			return user, nil
		}
	}
	return handlers.User{}, fmt.Errorf("User not found")
}
