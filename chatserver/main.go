package main

import (
	"fmt"
	"net/http"
	"chatserver/auth"


	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	http.HandleFunc("/", HomePage)
	http.HandleFunc("/ws", HandleConnections)
	http.HandleFunc("/login", auth.LoginHandler)
	http.HandleFunc("/register", auth.RegHandler)

	go HandleMessages()

	fmt.Println("Start server at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("Error starting server " + err.Error())
	}
}
