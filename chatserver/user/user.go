package user

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"
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
	filePath := "data/users.json"

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	var users []User

	if _, err := os.Stat(filePath); err == nil {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &users); err != nil {
				return errors.New("Existing users.json is not a valid JSON array")
			}
		}
	}

	users = append(users, user)

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	return nil
}
