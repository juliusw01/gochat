package auth

import(
	"encoding/base64"
    "encoding/json"
    "strings"
    "fmt"
)

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