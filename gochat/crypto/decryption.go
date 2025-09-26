package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const (
	service     = "gochat"
	username    = "gochat-enc"
	passwordLen = 32
)

func DecryptMessage(encodedCiphertext string, encodedNonce string, encodedAESKey string, user string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", fmt.Errorf("Error base64 decoding ciphertext %v", err)
	}
	nonce, err := base64.StdEncoding.DecodeString(encodedNonce)
	if err != nil {
		return "", fmt.Errorf("Error base64 decoding nonce %v", err)
	}
	encAESKey, err := base64.StdEncoding.DecodeString(encodedAESKey)
	if err != nil {
		return "", fmt.Errorf("Error base64 decoding aesKey %v", err)
	}
	aesKey := decryptAESKey(encAESKey, user)
	plaintext, err := AESDecrypt(ciphertext, nonce, aesKey)
	if err != nil {
		return "", fmt.Errorf("Error decrypting message %v", err)
	}
	return plaintext, nil
}

func decryptAESKey(encAESKey []byte, user string) []byte {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	privateKeyPath := filepath.Join(homeDir, ".gochat", user, "private.pem")

	pass, err := keyring.Get(service, username)
	if err != nil {
		log.Fatalf("Error retrieving passphrase for private key %v", err)
	}
	privateKeyPem, err := os.ReadFile(privateKeyPath)
	block, _ := pem.Decode(privateKeyPem)
	if block == nil {
		log.Fatal("Failed to decode PEM block")
	}

	decryptedDER, err := x509.DecryptPEMBlock(block, []byte(pass))
	if err != nil {
		log.Fatalf("Error decrypting PEM %v", err)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(decryptedDER)
	if err != nil {
		log.Fatalf("Error unmarshalling private key from PEM %v", err)
	}

	hash := sha512.New()
	aesKey, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, encAESKey, nil)
	if err != nil {
		log.Fatalf("Error decrypting AES key for message %v", err)
	}
	return aesKey
}
