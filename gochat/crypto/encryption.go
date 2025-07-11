package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/zalando/go-keyring"
)

const (
	service     = "gochat"
	username    = "gochat-enc"
	passwordLen = 32
)

func EncryptMessage(message string) string {

	homeDir, err := os.UserHomeDir()
	if err!= nil {
		fmt.Println(err)
	}

	dir := homeDir + "/.gochat/private.pem"

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = nil
		privateKey, publicKey, err := GenerateRSAKeys()
		if err != nil {
			log.Fatalf("RSA key pair could not be generated %v", err)
		}

		passphrase := getOrCreatePassphraseFromKeychain()
		encryptedPEM := encryptPrivateKeyToPEM(privateKey, passphrase)
		os.WriteFile(dir, encryptedPEM, 0600)

		uploadPublicKeyToServer(publicKey)
	}

	aesKey := make([]byte, 32)
	_, err = rand.Read(aesKey)
	if err != nil {
		log.Fatalf("Failed to generate random AES key %v", err)
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		log.Fatalf("Failed to create new AES key %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatalf("Failed to create Galois Counter Mode %v", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		log.Fatalf("Failed to generate nonce %v", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(message), nil)
	//encryptedAES := ""

	return string(ciphertext)
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
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}
	pemData := pem.EncodeToMemory(pemBlock)
	return pemData
}

func uploadPublicKeyToServer(publicKey *rsa.PublicKey) {
	keyBytes := x509.MarshalPKCS1PublicKey(publicKey)
	pemBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: keyBytes,
	}
	pemData := pem.EncodeToMemory(pemBlock)

	req, err := http.NewRequest("POST", "http://localhost:8080/upload/public-key", bytes.NewBuffer(pemData))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err!= nil {
		fmt.Println(err)
	}
	token, err := os.ReadFile(homeDir + "/.gochat/authToken.txt")
	if err != nil {
		fmt.Println("Error finding authToken. Please athenticate via 'gochat authenticate -u <username> -p <password>' first", err)
		return
	}
	jwtToken := string(token)

	req.Header.Set("Authorization", "Bearer " + jwtToken)
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
