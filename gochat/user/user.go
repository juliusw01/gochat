package user

import "time"

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
