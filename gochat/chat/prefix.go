package chat

import (
	"fmt"
	"gochat/crypto"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func CheckPrefix(currentRoom string, text string, username string, conn *websocket.Conn, recipient string) (int, string, string) {
	if strings.HasPrefix(text, "/join ") {
		currentRoom = strings.TrimSpace(strings.TrimPrefix(text, "/join "))
		recipient := ""
		fmt.Printf("---------- [%s] ----------", currentRoom)
		fmt.Printf("\n")
		msg := Message{
			Username: username,
			Message:  fmt.Sprintf("%s joined the room.", username),
			Room:     currentRoom,
			Sent:     time.Now(),
			Type:     "system",
		}
		conn.WriteJSON(msg)
		return 0, currentRoom, recipient
	}
	if strings.HasPrefix(text, "/exit") {
		msg := Message{
			Username: username,
			Message:  fmt.Sprintf("%s left the room.", username),
			Room:     currentRoom,
			Sent:     time.Now(),
			Type:     "system",
		}
		conn.WriteJSON(msg)
		return 1, currentRoom, ""
	}
	if strings.HasPrefix(text, "/dm") {
		currentRoom = ""
		parts := strings.SplitN(text, " ", 3)
		if len(parts) < 2 {
			fmt.Println("Usage: /dm <receipient> [<message>]")
			return 0, currentRoom, recipient
		}
		if recipient != parts[1] {
			
		}
		recipient := parts[1]
		fmt.Printf("---------- [%s] ----------", recipient)
		fmt.Printf("\n")
		if len(parts) == 2 {
			return 0, currentRoom, recipient
		}
		message := parts[2]
		encryptedMsg, nonce, aesKey := crypto.EncryptMessage(message, username, recipient)
		msg := Message{
			Username:  username,
			Message:   encryptedMsg,
			Sent:      time.Now(),
			Recipient: recipient,
			Nonce:     nonce,
			AESKey:    aesKey,
			Type:      "chat",
		}
		conn.WriteJSON(msg)
		return 0, currentRoom, recipient
	}
	return -1, currentRoom, recipient
}
