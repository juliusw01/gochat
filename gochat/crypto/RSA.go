package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func GenerateRSAKeys() (*rsa.PrivateKey, *rsa.PublicKey, error){
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		fmt.Print(err)
		return nil, nil, err
	}
	publicKey := &privateKey.PublicKey

	return privateKey, publicKey, nil	
}

func CreateRSAPair(user string){
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
	}

	privateKeyPath := filepath.Join(homeDir, ".gochat", user, "private.pem")

	//if no private key exists, create a new one
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		err = nil
		privateKey, publicKey, err := GenerateRSAKeys()
		if err != nil {
			log.Fatalf("RSA key pair could not be generated %v", err)
		}

		passphrase := getOrCreatePassphraseFromKeychain()
		encryptedPEM := encryptPrivateKeyToPEM(privateKey, passphrase)
		os.WriteFile(privateKeyPath, encryptedPEM, 0600)

		uploadPublicKeyToServer(publicKey, user)
	}
}

