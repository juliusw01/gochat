package auth

import (
	"chatserver/user"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func RegHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newUser user.User
	json.NewDecoder(r.Body).Decode(&newUser)

	username := strings.ToLower(newUser.Username)

	if username == "admin" {
		http.Error(w, "Username cannot be admin", http.StatusForbidden)
		return
	}

	existingUser, err := getUserFromDatabase(newUser.Username)
	if existingUser != nil {
		http.Error(w, "User already exists. Please select a different username", http.StatusConflict)
		return
	}
	if err != nil {
		log.Println("Error saving user.", err)
		http.Error(w, "User could not be saved", http.StatusInternalServerError)
		return
	}

	if err := user.SaveUser(newUser); err != nil {
		log.Println("Error saving user.", err)
		http.Error(w, "User could not be saved", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("User created successfully"))
}
