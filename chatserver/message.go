package main

import (
	"encoding/json"
	"os"
	"time"
)

type Message struct {
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	Room      string    `json:"room,omitempty"`
	Sent      time.Time `json:"sent"`
	Recipient string    `json:"recipient,omitempty"`
	Nonce     string    `json:"nonce,omitempty"`
	AESKey    string    `json:"aesKey,omitempty"`
}

func SaveMessage(messages []Message) error {
	//TODO: messages get overwritten everytime server restarts. Also data model is basically not existent. Don't save all messages in one JSON. Maybe per Chat?
	data, err := json.MarshalIndent(messages, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile("data/messages.json", data, 0644)
}
