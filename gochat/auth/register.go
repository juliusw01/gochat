package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gochat/user"
	"io/ioutil"
	"log"
	"net/http"
)

func RegUser(username string, password string) {

	//TODO: call function to create RSA key pair here!!!

	toReg := user.NewUser(username, password)
	userJson, err := json.Marshal(toReg)
	if err != nil {
		log.Fatalf("Error encoding Json: %v", err)
	}

	resp, err := http.Post("http://raspberrypi.fritz.box:8080/register", "application/json", bytes.NewBuffer(userJson))
	if err != nil {
		log.Fatalf("Error registering user %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading responde %v", err)
	}

	fmt.Println(string(body))
}
