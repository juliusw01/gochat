package main
import (
	"time"	
	"encoding/json"
    "os"
)


type Message struct {
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	Room      string    `json:"room,omitempty"`
	Sent      time.Time `json:"sent"`
	Recipient string    `json:"recipient,omitempty"`
}

func SaveMessage(messages []Message) error {
	data, err := json.MarshalIndent(messages, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile("data/messages.json", data, 0644)
}
