package crypto

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Key struct {
	Username  string `json:"username"`
	PublicKey string `json:"publickey"`
}

func SavePublicKey(pubKey Key) error {
	filePath := "data/publickey.json"

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	var keys []Key

	// Read existing data if the file exists
	if _, err := os.Stat(filePath); err == nil {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &keys); err != nil {
				return errors.New("Existing publickey.json is not valid JSON array")
			}
		}
	}

	// Append the new key
	keys = append(keys, pubKey)

	// Marshal and write the full array
	data, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	return nil
}
