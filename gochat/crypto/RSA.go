package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
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

