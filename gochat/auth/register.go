package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gochat/user"
	"io/ioutil"
	"net/http"
)

func RegUser(username string, password string) {

	toReg := user.NewUser(username, password)
	userJson, err := json.Marshal(toReg)
	if err != nil {
		fmt.Println("Error encoding Json: ", err)		
		return
	}

	resp, err := http.Post("http://localhost:8080/register", "application/json", bytes.NewBuffer(userJson))
	if err != nil {
		fmt.Println("Error registering user", err)
		return
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading responde", err)		
	}

	fmt.Println(string(body))
}