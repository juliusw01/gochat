package crypto

import (
	"chatserver/auth"
	"io"
	"log"
	"net/http"
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

}
