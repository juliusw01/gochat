package auth

import (
	"fmt"
	"time"
	"encoding/base64"
    "encoding/json"
    "strings"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("secret-key")

func CreateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, 
        jwt.MapClaims{ 
        "username": username, 
        "exp": time.Now().Add(time.Hour * 24).Unix(), 
        })

    tokenString, err := token.SignedString(secretKey)
    if err != nil {
    	return "", err
    }

	return tokenString, nil
}

func VerifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
	   return secretKey, nil
	})

	if err != nil {
	   return err
	}

	if !token.Valid {
	   return fmt.Errorf("invalid token")
	}
	return nil
}

func ExtractUserFromToken(token string) (string, error){
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format")
	}

	payload := parts[1]

	if m := len(payload) % 4; m != 0 {
        payload += strings.Repeat("=", 4-m)
    }

    decoded, err := base64.URLEncoding.DecodeString(payload)
    if err != nil {
        return "", fmt.Errorf("error decoding payload: %w", err)
    }

    var claims map[string]interface{}
    if err := json.Unmarshal(decoded, &claims); err != nil {
        return "", fmt.Errorf("error unmarshaling JSON: %w", err)
    }

    username, ok := claims["username"].(string)
    if !ok {
        return "", fmt.Errorf("username claim not found or invalid")
    }

    return username, nil
}