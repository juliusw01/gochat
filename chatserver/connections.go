package main

import (
	"chatserver/auth"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var clientsMutex sync.Mutex
var broadcast = make(chan Message)
var rooms = make(map[string]map[*websocket.Conn]bool)
var roomsMutex sync.Mutex
var messages []Message
var userConnections = make(map[string]*websocket.Conn)
var connToUsername = make(map[*websocket.Conn]string)
var userConnMutex sync.Mutex

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Missing authorization header")
		return
	}
	tokenString = tokenString[len("Bearer "):]
	err := auth.VerifyToken(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Invalid token")
		return
	}
	username, err := auth.ExtractUserFromToken(tokenString)
	if err != nil {
		fmt.Errorf("Error extracting username from token %v", err)
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		userConnMutex.Lock()
		user, ok := connToUsername[conn]
		if ok && userConnections[user] == conn {
			delete(userConnections, user)
		}
		delete(connToUsername, conn)
		userConnMutex.Unlock()
		conn.Close()
	}()

	var currentRoom string

	clients[conn] = true

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println(err)
			clientsMutex.Lock()
			delete(clients, conn)
			clientsMutex.Unlock()
			if currentRoom != "" {
				roomsMutex.Lock()
				delete(rooms[currentRoom], conn)
				roomsMutex.Unlock()
			}
			return
		}

		//username = msg.Username
		if username != msg.Username {
			fmt.Errorf("Username in Message does not match with username in auth token")
			continue
		}

		if msg.Room == "dm" {
			invalidRoom := Message{
				Username:  "Admin",
				Message:   fmt.Sprintf("Invalid room name. Name cannot be named 'dm'. You are still in your current room: %s", currentRoom),
				Sent:      time.Now(),
				Recipient: username,
			}
			err = conn.WriteJSON(invalidRoom)
			if err != nil {
				fmt.Println(err)
			}
			continue
		}

		userConnMutex.Lock()
		if _, ok := userConnections[username]; !ok {
			userConnections[username] = conn
		}

		oldConn, exists := userConnections[username]
		if exists && oldConn != conn {
			oldConn.Close()
		}

		userConnections[username] = conn
		connToUsername[conn] = username
		userConnMutex.Unlock()

		if currentRoom != msg.Room {
			roomsMutex.Lock()
			if currentRoom != "" {
				delete(rooms[currentRoom], conn)
			}
			if rooms[msg.Room] == nil {
				rooms[msg.Room] = make(map[*websocket.Conn]bool)
			}
			rooms[msg.Room][conn] = true
			roomsMutex.Unlock()
			currentRoom = msg.Room
		}
		broadcast <- msg
		messages = append(messages, msg)
		err = SaveMessage(messages)
		if err != nil {
			fmt.Println("Error saving messages:")
			fmt.Println(err)
		}

		if len(rooms[currentRoom]) == 0 {
			delete(rooms, currentRoom)
		}
	}
}

func HandleMessages() {
	for {
		msg := <-broadcast

		//send direct message to user based on username
		if msg.Recipient != "" {
			userConnMutex.Lock()
			receipientConn, ok := userConnections[msg.Recipient]
			userConnMutex.Unlock()

			if ok {
				err := receipientConn.WriteJSON(msg)
				if err != nil {
					fmt.Printf("Error sending DM: ")
					fmt.Println(err)
					receipientConn.Close()
				}
			}
			continue
		}

		//send broadcast message to people in chat room
		roomsMutex.Lock()
		clients := rooms[msg.Room]
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				fmt.Println(err)
				client.Close()
				rooms[msg.Room][client] = false
				delete(rooms[msg.Room], client)

				clientsMutex.Lock()
				delete(clients, client)
				clientsMutex.Unlock()
			}
		}
		roomsMutex.Unlock()
	}
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	b, err := os.ReadFile("homepage.txt")
	if err != nil {
		fmt.Println(err)
	}
	w.Write(b)
}
