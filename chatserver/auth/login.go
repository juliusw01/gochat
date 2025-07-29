package auth

import (
	"chatserver/user"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var user user.User
	json.NewDecoder(r.Body).Decode(&user)

	actual, err := getUserFromDatabase(user.Username)
	if err != nil {
		http.Error(w, "Could not get user from database", http.StatusUnauthorized)
	}

	//TODO: Implement actual users
	if user.Username == actual.Username && user.Password == actual.Password {
		tokenString, err := CreateToken(user.Username)
		if err != nil {
			http.Error(w, "Could not create auth token", http.StatusInternalServerError)
			fmt.Errorf("No username found")
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, tokenString)
		return
		//} else if user.Username == "Chek1" && user.Password == "123456" {
		//	tokenString, err := CreateToken(user.Username)
		//	if err != nil {
		//		w.WriteHeader(http.StatusInternalServerError)
		//		fmt.Errorf("No username found")
		//	}
		//	w.WriteHeader(http.StatusOK)
		//	fmt.Fprint(w, tokenString)
		//	return
	} else {
		http.Error(w, "Username or password incorrect", http.StatusUnauthorized)
		fmt.Fprint(w, "Invalid credentials")
	}
}

func getUserFromDatabase(username string) (*user.User, error) {
	filePath := "data/users.json"

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return nil, err
	}

	var users []user.User

	if _, err := os.Stat(filePath); err == nil {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &users); err != nil {
				return nil, errors.New("Existing users.json is not a valid JSON array")
			}
		}
	}

	for i := range users {
	if users[i].Username == username {
		return &users[i], nil
		}
	}

	return nil, errors.New("User does not exist")
}
