package crypto

import (
	"chatserver/auth"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

func PublicKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Missing authorization header")
		return
	}
	tokenString = tokenString[len("Bearer "):]
	errr := auth.VerifyToken(tokenString)
	if errr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Invalid token")
		return
	}

	body, _ := io.ReadAll(r.Body)

	b64Data := base64.StdEncoding.EncodeToString(body)
	//fmt.Print(b64Data)
	//TODO: extract the username from jwt and save the username + public key

}
