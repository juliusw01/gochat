package crypto

import (
	"chatserver/auth"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func PublicKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	err := auth.Authenticate(w, r)
	if err != nil {
		return
	}

	username, err := auth.ExtractUserFromToken(r.Header.Get("Authorization"))
	if err != nil {
		log.Fatalf("Error extracting username from AuthToken %v", err)
		http.Error(w, "Error extracting username from AuthToken", http.StatusInternalServerError)
	}
	buf := new(strings.Builder)
	_, err = io.Copy(buf, r.Body)
	if err != nil {
		log.Fatalf("Error retrieving public key from request %v", err)
		http.Error(w, "Error retrieving public key from request", http.StatusInternalServerError)
	}

	pubKey := buf.String()

	key := Key{username, pubKey}
	err = SavePublicKey(key)
	if err != nil {
		log.Fatalf("Error saving public key %v", err)
		http.Error(w, "Error saving public key", http.StatusInternalServerError)
	}
}

func GetPublicKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := auth.Authenticate(w, r)
	if err != nil {
		return
	}

	recipient := r.PathValue("recipient")
	_, s, _ := checkIfPubKeyExists(recipient)
	w.Write([]byte(s))

}

func checkIfPubKeyExists(username string) (bool, string, error) {
	// Read the JSON file
	jsonKeys, err := os.ReadFile("data/publickey.json")
	if err != nil {
		return false, "", fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal JSON into slice of Key
	var keys []Key
	if err := json.Unmarshal(jsonKeys, &keys); err != nil {
		return false, "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Search for username
	for _, k := range keys {
		if k.Username == username {
			return true, k.PublicKey, nil
		}
	}

	return false, "", nil
}
