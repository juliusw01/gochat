package auth

import (
	"chatserver/user"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func RegHandler(w http.ResponseWriter, r *http.Request){
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
	
	if err := user.SaveUser(newUser); err != nil {
		fmt.Println("Error saving user.", err)
		http.Error(w, "User could not be saved", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("User created successfully"))
}