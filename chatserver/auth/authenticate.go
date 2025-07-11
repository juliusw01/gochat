package auth

import (
	"fmt"
	"net/http"
)

func Authenticate(w http.ResponseWriter, r *http.Request) error {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Missing authorization header")
		return fmt.Errorf("Missing authorization header")
	}
	tokenString = tokenString[len("Bearer "):]
	errr := VerifyToken(tokenString)
	if errr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Invalid token")
		return fmt.Errorf("Invalid token")
	}
	return nil
}
