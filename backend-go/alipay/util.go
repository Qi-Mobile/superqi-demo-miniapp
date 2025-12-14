package alipay

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Decode PEM block
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing private key")
	}

	// Parse the private key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format if PKCS1 fails
		if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
			if privateKey, ok := key.(*rsa.PrivateKey); ok {
				return privateKey, nil
			}
			return nil, errors.New("key is not an RSA private key")
		}
		return nil, err
	}

	return privateKey, nil
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Decode PEM block
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing public key")
	}

	// Parse the public key
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("key is not an RSA public key")
	}

	return publicKey, nil
}
