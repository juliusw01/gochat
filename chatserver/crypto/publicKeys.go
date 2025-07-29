package crypto

import (
	"chatserver/auth"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

//TODO: Remove ALL log.Fatalf() calls from chatserver
//Fatalf() calls os.Exit(1), which ends the program – this is unwanted behaviour for the server. The server needs to keep running

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
		//log.Fatalf("Error extracting username from AuthToken %v", err)
		http.Error(w, "Error extracting username from AuthToken", http.StatusInternalServerError)
	}
	buf := new(strings.Builder)
	_, err = io.Copy(buf, r.Body)
	if err != nil {
		//log.Fatalf("Error retrieving public key from request %v", err)
		http.Error(w, "Error retrieving public key from request", http.StatusInternalServerError)
	}

	pubKey := buf.String()

	key := Key{username, pubKey}
	alreadyExists, _, err := checkIfPubKeyExists(username)
	if err != nil {
		//log.Fatalf("Error checking if public key already exists for user %v", err)
		http.Error(w, "Error checking if public key already exists for user", http.StatusInternalServerError)
	}
	// If a user registers and creates a private/public key pair, only one public key is allowed for a user --> prevent identity theft
	//TODO: What happens if a user misplaces/deletes the private key or changes devices. Implement a method to allow new or even multiple public keys (for multi device functionality)
	if alreadyExists {
		http.Error(w, "User already has a public key", http.StatusForbidden)
	}
	err = SavePublicKey(key)
	if err != nil {
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
	_, s, err := checkIfPubKeyExists(recipient)
	if err != nil {
		http.Error(w, "Error checking for recipient's public key", http.StatusInternalServerError)
	}
	w.Write([]byte(s))

}

func checkIfPubKeyExists(username string) (bool, string, error) {
	// Read the JSON file
	jsonKeys, err := os.ReadFile("data/publickey.json")
	if err != nil {
		return false, "", fmt.Errorf("failed to read file: %v", err)
	}

	// Unmarshal JSON into slice of Key
	var keys []Key
	if err := json.Unmarshal(jsonKeys, &keys); err != nil {
		return false, "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Search for username
	for _, k := range keys {
		if k.Username == username {
			return true, k.PublicKey, nil
		}
	}

	return false, "", nil
}
