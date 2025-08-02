//TODO: Remove this class and implement in connections.go

package call

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
	upgrader  = websocket.Upgrader{}
)

func CallSignalMessage(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	log.Println("New client connected")

	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Relay to all other clients
		clientsMu.Lock()
		for c := range clients {
			if c != conn {
				c.WriteMessage(mt, msg)
			}
		}
		clientsMu.Unlock()
	}

	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()

	log.Println("Client disconnected")
}
