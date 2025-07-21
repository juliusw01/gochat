package crypto

import (
	"encoding/base64"
	"fmt"
)

func DecryptMessage(encodedCiphertext string, encodedNonce string, encodedAESKey string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", fmt.Errorf("Error base64 decoding ciphertext %w", err)
	}
	nonce, err := base64.StdEncoding.DecodeString(encodedNonce)
	if err != nil {
		return "", fmt.Errorf("Error base64 decoding nonce %w", err)
	}
	aesKey, err := base64.StdEncoding.DecodeString(encodedAESKey)
	if err != nil {
		return "", fmt.Errorf("Error base64 decoding aesKey %w", err)
	}
	//TODO: When AES key will be encrypted with public key, it has to be decrypted with private key first
	plaintext, err := AESDecrypt(ciphertext, nonce, aesKey)
	if err != nil {
		return "", fmt.Errorf("Error decrypting message %w", err)
	}
	return plaintext, nil
}
