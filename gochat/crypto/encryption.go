package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

//const (
//	service     = "gochat"
//	username    = "gochat-enc"
//	passwordLen = 32
//)

func EncryptMessage(message string, user string, recipient string) (string, string, string) {

	ciphertext, aesKey, nonce, err := EncryptAES(message)
	if err != nil {
		log.Fatalf("Could not AES encrypt message %w", err)
	}

	encryptedAESKey := encryptAESKey(aesKey, recipient, user)

	encodedCiphertext := base64.StdEncoding.EncodeToString(ciphertext)
	encodedNonce := base64.StdEncoding.EncodeToString(nonce)
	encodedAESKey := base64.StdEncoding.EncodeToString(encryptedAESKey)

	return encodedCiphertext, encodedNonce, encodedAESKey

}

func encryptAESKey(aesKey []byte, recipient string, user string) []byte {
	req, err := http.NewRequest("GET", "http://raspberrypi.fritz.box:8080/public-key/"+recipient, nil)

	if err != nil {
		log.Fatalf("Error retrieving public key for recipient %w", err)
		return []byte("")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
	}
	tokenDir := filepath.Join(homeDir, ".gochat", user, "authToken.txt")
	token, err := os.ReadFile(tokenDir)
	if err != nil {
		fmt.Println("Error finding authToken. Please athenticate via 'gochat authenticate -u <username> -p <password>' first", err)
		return []byte("")
	}
	jwtToken := string(token)

	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Content-Type", "application/x-pem-file")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Server responded with status: %s", resp.Status)
	}

	pubKeyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Could not read public key from response body %w", err)
	}
	if pubKeyBytes == nil {
		log.Fatal("Empty response body")
	}
	
	block, _ := pem.Decode(pubKeyBytes)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		log.Fatal("Failed to decode PEM block containing public key")
		return []byte("")
	}
	pubKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		log.Fatalf("Could not parse DER encoded public key %w", err)
	}
	//pubKey, isRSAKey := pubKeyInterface.(*rsa.PublicKey)
	//if !isRSAKey {
	//	log.Fatalf("Public key parsed is not an RSA public key %w", err)
	//}

	hash := sha512.New()
	encAESKey, err := rsa.EncryptOAEP(hash, rand.Reader, pubKey, aesKey, nil)

	return encAESKey
}

func getOrCreatePassphraseFromKeychain() string {

	pass, err := keyring.Get(service, username)
	if err == nil {
		return pass
	}

	randBytes := make([]byte, passwordLen)
	_, err = rand.Read(randBytes)
	if err != nil {
		log.Fatalf("Failed to generate random password %v", err)
	}

	newPass := base64.StdEncoding.EncodeToString(randBytes)
	err = keyring.Set(service, username, newPass)

	if err != nil {
		log.Fatalf("failed to store passphrase in keychain: %v", err)
	}

	return newPass
}

func encryptPrivateKeyToPEM(privateKey *rsa.PrivateKey, passphrase string) []byte {
	keyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	// Encrypt the PEM block using a passphrase from keyring
	block, err := x509.EncryptPEMBlock(
		rand.Reader,
		"RSA PRIVATE KEY",
		keyBytes,
		[]byte(passphrase),
		x509.PEMCipherAES256,
	)
	if err != nil {
		log.Fatalf("Failed to encrypt private key: %v", err)
	}

	// Encode to PEM format
	pemData := pem.EncodeToMemory(block)
	return pemData
}

func uploadPublicKeyToServer(publicKey *rsa.PublicKey, user string) {
	keyBytes := x509.MarshalPKCS1PublicKey(publicKey)
	pemBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: keyBytes,
	}
	pemData := pem.EncodeToMemory(pemBlock)

	req, err := http.NewRequest("POST", "http://raspberrypi.fritz.box:8080/upload/public-key", bytes.NewBuffer(pemData))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
	}
	tokenDir := filepath.Join(homeDir, ".gochat", user, "authToken.txt")
	token, err := os.ReadFile(tokenDir)
	if err != nil {
		fmt.Println("Error finding authToken. Please athenticate via 'gochat authenticate -u <username> -p <password>' first", err)
		return
	}
	jwtToken := string(token)

	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Content-Type", "application/x-pem-file")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Server responded with status: %s", resp.Status)
	}

}
