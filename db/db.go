package db

import (
	"fmt"
	"journalCli/handlers"
)

func GetUserByID(id string) (handlers.User, error) {
	for _, user := range handlers.Users {
		if user.Id == id {
			return user, nil
		}
	}
	return handlers.User{}, fmt.Errorf("User not found")
}
