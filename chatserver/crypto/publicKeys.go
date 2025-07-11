package crypto

import (
	"chatserver/auth"
	"log"
	"net/http"
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
	}
	key := Key{username, "someKey"}
	SavePublicKey(key)
}

func GetPublicKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

}
