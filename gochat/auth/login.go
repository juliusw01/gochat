package auth

import (
	"bytes"
	"encoding/json"
	"gochat/user"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func UserLogin(username string, password string) {
	user := user.User{
		Username: username,
		Password: password,
	}
	userJson, err := json.Marshal(user)
	if err != nil {
		log.Fatalf("Error encoding Json: %w", err)
		return
	}
	resp, err := http.Post("http://localhost:8080/login", "application/json", bytes.NewBuffer(userJson))
	if err != nil {
		log.Fatalf("Error making login request: %w", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %w", err)
	}
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	dir = dir + "/.gochat/" + username + "/"
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = os.WriteFile(dir+"authToken.txt", body, 0600)
	if err != nil {
		log.Fatal(err)
	}
}
