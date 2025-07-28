package crypto

import (
	"encoding/json"
	"errors"
	"io/ioutil"
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
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &keys); err != nil {
				return errors.New("existing publickey.json is not valid JSON array")
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

	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	return nil
}
