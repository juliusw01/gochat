package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gochat/user"
	"io/ioutil"
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
		fmt.Println("Error encoding Json: ", err)		
		return
	}
	resp, err := http.Post("http://localhost:8080/login", "application/json", bytes.NewBuffer(userJson))
	if err != nil {
		fmt.Println("Error making login request", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading responde", err)		
	}
	dir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
	}
	dir = dir + "/.gochat/"
	err = os.MkdirAll(dir, 0700)
    if err != nil {
		fmt.Println(err)
        return 
    }
	err = os.WriteFile(dir + "authToken.txt", body, 0600)
	if err != nil {
		fmt.Println(err)
	}
}