package chat

import (
	"fmt"
	"gochat/crypto"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func CheckPrefix(currentRoom string, text string, username string, conn *websocket.Conn) (int, string) {
	if strings.HasPrefix(text, "/join ") {
		currentRoom = strings.TrimSpace(strings.TrimPrefix(text, "/join "))
		fmt.Printf("---------- [%s] ----------", currentRoom)
		fmt.Printf("\n")
		msg := Message{
			Username: username,
			Message:  fmt.Sprintf("%s joined the room.", username),
			Room:     currentRoom,
			Sent:     time.Now(),
		}
		conn.WriteJSON(msg)
		return 0, currentRoom
	}
	if strings.HasPrefix(text, "/exit") {
		msg := Message{
			Username: username,
			Message:  fmt.Sprintf("%s left the room.", username),
			Room:     currentRoom,
			Sent:     time.Now(),
		}
		conn.WriteJSON(msg)
		return 1, currentRoom
	}
	if strings.HasPrefix(text, "/dm") {
		parts := strings.SplitN(text, " ", 3)
		if len(parts) < 3 {
			fmt.Println("Usage: /dm <receipient> <message>")
			return 0, currentRoom
		}
		recipient := parts[1]
		message := parts[2]
		encryptedMsg := crypto.EncryptMessage(message)
		fmt.Println(encryptedMsg)
		msg := Message{
			Username: username,
			//Message:   encryptedMsg,
			Message:   message,
			Sent:      time.Now(),
			Recipient: recipient,
		}
		conn.WriteJSON(msg)
		return 0, currentRoom
	}
	return -1, currentRoom
}
