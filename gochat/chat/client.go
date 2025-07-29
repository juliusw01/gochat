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
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

var rooms = make(map[string]map[*websocket.Conn]bool)

//var roomsMutex sync.Mutex

func StartClient(user string) {
	currentRoom := "general"
	recipient := ""

	reader := bufio.NewReader(os.Stdin)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	/**Username is passed as an argument to support multiple accounts on one client (primarily for testing purposes)
	* Bad for user experience --> instead of 'gochat chat' username has to be passed too 'gochat chat -u Chek'
	* TODO: Either remove multi user support or find a better way
	**/
	tokenDir := filepath.Join(homeDir, ".gochat", user, "authToken.txt")
	token, err := os.ReadFile(tokenDir)
	if err != nil {
		log.Fatalf("Error finding authToken. Please athenticate via 'gochat authenticate -u <username> -p <password>' first: %v", err)
		return
	}
	tokenString := string(token)
	//Eventhough we pass the username as an arg, we want to extract the username from the signed token
	username, err := auth.ExtractUserFromToken(tokenString)
	if err != nil {
		log.Fatalf("User cannot be extracted from auth token: %v", err)
	}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+tokenString)
	conn, _, err := websocket.DefaultDialer.Dial("ws://raspberrypi.fritz.box:8080/ws", header)
	if err != nil {
		log.Fatalf("Dial error: %v", err)
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
				messageText, err = crypto.DecryptMessage(msg.Message, msg.Nonce, msg.AESKey, username)
				if err != nil {
					log.Fatalf("Message could not be decrypted %v", err)
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

		//TODO: Set fix prefix --> user has to set '/dm <user> <message>' every time. Treat dm's like chatrooms and save them. Prefix should only be specified when smth changes
		var i int
		i, currentRoom, recipient = CheckPrefix(currentRoom, text, username, conn, recipient)
		if i == 1 {
			fmt.Println("\nDisconnected from server.")
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			conn.Close()
			os.Exit(0)
		} else if i == 0 {
			continue
		}

		if recipient != "" {
			encryptedMsg, nonce, aesKey := crypto.EncryptMessage(text, username, recipient)

			msg := Message{
				Username:  username,
				Message:   encryptedMsg,
				Sent:      time.Now(),
				Recipient: recipient,
				Nonce:     nonce,
				AESKey:    aesKey,
			}

			err = conn.WriteJSON(msg)
			if err != nil {
				log.Println("Write error:", err)
				return
			}
		} else {
			msg := Message{
				Username:  username,
				Message:   text,
				Room:      currentRoom,
				Sent:      time.Now(),
				Recipient: recipient,
			}

			err = conn.WriteJSON(msg)
			if err != nil {
				log.Println("Write error:", err)
				return
			}
		}

	}
}

func initMessage(conn *websocket.Conn, username string) {
	msg := Message{
		Username: username,
		Message:  fmt.Sprintf("%s joined the chat.", username),
		Room:     "general",
		Sent:     time.Now(),
	}

	err := conn.WriteJSON(msg)
	if err != nil {
		log.Println("Write error:", err)
		return
	}
}
