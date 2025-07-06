package user

import (
	"encoding/json"
	"os"
	"time"
)

type User struct {
	Username string    `json:"username"`
	Password string    `json:"password"`
	Created  time.Time `json:"created"`
}

func NewUser(username string, password string) User {
	return User{
		Username: username,
		Password: password,
		Created:  time.Now(),
	}
}

func SaveUser(user User) error {
	data, err := json.MarshalIndent(user, "", " ")
	if err != nil {
		return err
	}
	f, err := os.OpenFile("data/users.json", os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return err
	}

	return nil
}
