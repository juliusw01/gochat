package chat

import "time"

type Message struct {
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	Room      string    `json:"room,omitempty"`
	Sent      time.Time `json:"sent"`
	Recipient string    `json:"recipient,omitempty"`
	Nonce     string    `json:"nonce,omitempty"`
	AESKey    string    `json:"aesKey,omitempty"`
}
