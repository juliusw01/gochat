// chat/sessions.go or at the top of chat package
package chat

import (
	"gochat/call"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

var activeCalls = struct {
	sync.Mutex
	sessions map[string]*call.CallSession
}{sessions: make(map[string]*call.CallSession)}

// get or create a call session
func getOrCreateCallSession(user, peer string, conn *websocket.Conn) *call.CallSession {
	activeCalls.Lock()
	defer activeCalls.Unlock()

	if s, ok := activeCalls.sessions[peer]; ok {
		return s
	}

	sess, err := call.NewCallSession(user, conn, peer)
	if err != nil {
		log.Println("Error creating call session:", err)
		return nil
	}
	activeCalls.sessions[peer] = sess
	return sess
}

// remove session when done
func removeCallSession(peer string) {
	activeCalls.Lock()
	defer activeCalls.Unlock()
	delete(activeCalls.sessions, peer)
}
