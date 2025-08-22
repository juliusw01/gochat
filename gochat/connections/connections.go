package connections

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gochat/auth"

	"github.com/gorilla/websocket"
)

func GetConnection(user string) (*websocket.Conn, error) {
	token := Authenticate(user)

	username, err := auth.ExtractUserFromToken(token)
	if err != nil {
		log.Fatalf("Cannot extract username from token: %v", err)
	}

	if username != user {
		log.Fatalf("Entered username does not match username in access token")
	}

	conn, err := ConnectToServer(token)
	if err != nil {
		log.Printf("Connection failed: %v", err)
		return nil, err
	}

	return conn, nil
}

func ConnectToServer(token string) (*websocket.Conn, error) {
	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)
	conn, resp, err := websocket.DefaultDialer.Dial("ws://raspberrypi.fritz.box:8080/ws", header)
	if err != nil {
		if resp.StatusCode == http.StatusUnauthorized {
			log.Fatalf("Invalid access token %v", resp.Status)
		}
		fmt.Println(resp.Status)
		return nil, err
	}
	return conn, nil
}

func Authenticate(user string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	tokenDir := filepath.Join(homeDir, ".gochat", user, "authToken.txt")
	token, err := os.ReadFile(tokenDir)
	if err != nil {
		log.Fatalf("Auth token not found. Please run: gochat authenticate -u <username> -p <password>")
	}
	return string(token)
}
