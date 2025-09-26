package chat

import (
	"bufio"
	"fmt"
	"gochat/auth"
	"gochat/call"
	"gochat/connections"
	"gochat/crypto"
	"gochat/misc"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/andybrewer/mack"
	"github.com/gorilla/websocket"
)

var currentInput string

func StartClient(user string, detach bool) {
	token := connections.Authenticate(user)

	username, err := auth.ExtractUserFromToken(token)
	if err != nil {
		log.Fatalf("Cannot extract username from token: %v", err)
	}

	// Handle shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down client...")
		os.Exit(0)
	}()

	// Main reconnect loop for daemon mode
	for {
		conn, err := connections.ConnectToServer(token)
		if err != nil {
			log.Printf("Connection failed: %v. Retrying in 5s...\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("Connected to server.")
		go receiveMessages(conn, username)
		initMessage(conn, username)
		startPingLoop(conn)

		if !detach {
			readFromStdinAndSend(conn, username)
			return // Exit after interactive session ends
		} else {
			select {} // Block forever until killed
		}
	}
}

func receiveMessages(conn *websocket.Conn, username string) {
	fmt.Println("Start receiving messages...")
	for {
		var msg call.Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("Read error:", err)
			return
		}

		switch msg.Type {
		case "chat":
			handleChatMessage(msg, username)
		case "system":
			//TODO: Build helper function to print messages properly. Especially when more message types exist
			fmt.Print("\r\033[K")
			log.Printf("[SYSTEM] %s\n", msg.Message)
			fmt.Printf("> %s", currentInput)
		case "offer", "answer", "candidate":
			sess := getOrCreateCallSession(username, msg.Username, conn)
			if sess == nil {
				log.Println("Failed to get or create call session")
				continue
			}
			if msg.Type == "offer" {
				//misc.Notify(msg.Username + " wants to call you", "gochat", "", "Blow.aiff")
				alert := mack.AlertOptions{
					Title:         "gochat",
					Message:       msg.Username + " is calling you",
					Duration:      15,
					Buttons:       "Yes, No",
					DefaultButton: "Yes",
				}
				resp, err := mack.AlertBox(alert)
				if err != nil {
					log.Println(err)
					continue
				}
				if resp.Clicked == "Yes" {
					misc.Notify("Call accepted", "gochat", "", "Blow.aiff")
					sess.AcceptCall()
				} else {
					continue
				}
			}
			sess.HandleSignalMessage(msg)
		default:
			fmt.Print("\r\033[K")
			log.Printf("[UNKNOWN TYPE] %v\n", msg)
			fmt.Printf("> %s", currentInput)
		}
	}
}

func handleChatMessage(msg call.Message, username string) {
	messageText := msg.Message
	receivedIn := msg.Room

	if msg.Recipient != "" { // DM
		receivedIn = "dm"
		plain, err := crypto.DecryptMessage(msg.Message, msg.Nonce, msg.AESKey, username)
		if err != nil {
			log.Printf("Failed to decrypt message: %v\n", err)
		} else {
			messageText = plain
		}
	}

	if msg.Username != username {
		// Move cursor to start of line and clear it
		fmt.Print("\r\033[K")
		// Print the incoming message
		fmt.Printf("%s [%s][%s]: %s\n",
			msg.Sent.Local().Format("2006-01-02 15:04:05"),
			receivedIn,
			msg.Username,
			messageText)
		// Restore user input
		fmt.Printf("> %s", currentInput)
		misc.Notify(msg.Username+" sent you a message", "gochat", "", "Blow.aiff")
	}
}

func readFromStdinAndSend(conn *websocket.Conn, username string) {
	currentRoom := "general"
	recipient := ""

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Input error:", err)
			continue
		}

		text = strings.TrimSpace(text)
		//TODO: currentInput must be updated on every keystroke, but it is not!
		currentInput = ""
		if text == "" {
			continue // Ignore empty messages
		}

		var cmd int
		cmd, currentRoom, recipient = CheckPrefix(currentRoom, text, username, conn, recipient)

		if cmd == 1 { // Exit
			log.Println("Disconnecting...")
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			conn.Close()
			os.Exit(0)
		} else if cmd == 0 {
			continue // Command handled internally
		}

		if recipient != "" {
			encrypted, nonce, aesKey := crypto.EncryptMessage(text, username, recipient)
			msg := call.Message{
				Username:  username,
				Message:   encrypted,
				Nonce:     nonce,
				AESKey:    aesKey,
				Recipient: recipient,
				Sent:      time.Now(),
				Type:      "chat",
			}
			conn.WriteJSON(msg)
		} else {
			msg := call.Message{
				Username: username,
				Message:  text,
				Room:     currentRoom,
				Sent:     time.Now(),
				Type:     "chat",
			}
			//TODO: Also encrypt group messages!
			/*
			* Get all members and their public key in a chat room from the server and encrypt the AES key with every public key
			*/
			conn.WriteJSON(msg)
		}
	}
}

func initMessage(conn *websocket.Conn, username string) {
	msg := call.Message{
		Username: username,
		Message:  fmt.Sprintf("%s joined the chat.", username),
		Room:     "general",
		Sent:     time.Now(),
		Type:     "system",
	}
	conn.WriteJSON(msg)
}

func startPingLoop(conn *websocket.Conn) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
				log.Println("Ping error:", err)
				return
			}
		}
	}()
}
