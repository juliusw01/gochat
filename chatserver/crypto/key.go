package crypto

import (
	"encoding/json"
	"os"
)

type Key struct {
	Username  string `json:"username"`
	PublicKey string `json:"publickey"`
}

func SavePublicKey(pubKey Key) error {
	data, err := json.MarshalIndent(pubKey, "", "")
	if err != nil {
		return err
	}
	f, err := os.OpenFile("data/publickey.json", os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return err
	}

	return nil
}
