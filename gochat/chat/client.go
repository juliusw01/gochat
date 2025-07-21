package chat

import (
	"bufio"
	"fmt"
	"gochat/auth"
	"gochat/crypto"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

var rooms = make(map[string]map[*websocket.Conn]bool)
var currentRoom string = "general"

//var roomsMutex sync.Mutex

func StartClient() {
	reader := bufio.NewReader(os.Stdin)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
	}
	token, err := os.ReadFile(homeDir + "/.gochat/authToken.txt")
	if err != nil {
		fmt.Println("Error finding authToken. Please athenticate via 'gochat authenticate -u <username> -p <password>' first", err)
		return
	}
	tokenString := string(token)
	username, err := auth.ExtractUserFromToken(tokenString)
	if err != nil {
		fmt.Println("User cannot be extracted from auth token", err)
	}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+tokenString)
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", header)
	if err != nil {
		log.Fatal("Dial error:", err)
	}
	defer conn.Close()

	initMessage(conn, username)

	// Receive messages in a goroutine
	go func() {
		for {
			var msg Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Println("Read error:", err)
				return
			}
			messageText := msg.Message
			received_in := msg.Room
			if msg.Recipient != "" {
				received_in = "dm"
				messageText, err = crypto.DecryptMessage(msg.Message, msg.Nonce, msg.AESKey)
				if err != nil {
					log.Fatalf("Message could not be decrypted %w", err)
				}
			}
			fmt.Printf("%s [%s][%s]: %s\n", msg.Sent.Format("2006-01-02 15:04:05"), received_in, msg.Username, messageText)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		fmt.Println("\nDisconnected from server.")
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		conn.Close()
		os.Exit(0)
	}()

	// Read from stdin and send
	for {
		//fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}
		text = strings.TrimSpace(text)

		var i int
		i, currentRoom = CheckPrefix(currentRoom, text, username, conn)
		if i == 1 {
			fmt.Println("\nDisconnected from server.")
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			conn.Close()
			os.Exit(0)
		} else if i == 0 {
			continue
		}

		msg := Message{
			Username: username,
			Message:  text,
			Room:     currentRoom,
			Sent:     time.Now(),
		}

		err = conn.WriteJSON(msg)
		if err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}

func initMessage(conn *websocket.Conn, username string) {
	msg := Message{
		Username: username,
		Message:  fmt.Sprintf("%s joined the chat.", username),
		Room:     currentRoom,
		Sent:     time.Now(),
	}

	err := conn.WriteJSON(msg)
	if err != nil {
		log.Println("Write error:", err)
		return
	}
}
