package auth

import (
	"bytes"
	"encoding/json"
	"gochat/user"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func UserLogin(username string, password string) {
	user := user.User{
		Username: username,
		Password: password,
	}
	userJson, err := json.Marshal(user)
	if err != nil {
		log.Fatalf("Error encoding Json: %v", err)
		return
	}

	resp, err := http.Post("http://raspberrypi.fritz.box:8080/login", "application/json", bytes.NewBuffer(userJson))
	if err != nil {
		//fmt.Println(resp.StatusCode)
		log.Fatalf("Error making login request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	if resp.StatusCode != 200 {
		log.Println(string(body))
		return
	}

	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	fileDir := filepath.Join(dir, ".gochat", username)
	err = os.MkdirAll(fileDir, 0700)
	if err != nil {
		log.Fatal(err)
		return
	}
	filePath := filepath.Join(fileDir, "authToken.txt")
	err = os.WriteFile(filePath, body, 0600)
	if err != nil {
		log.Fatal(err)
	}
}
