package jwe

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/square/go-jose/v3"
)

type TokenClaims struct {
	UserID      string `json:"user_id"`
	AccessToken string `json:"access_token"`
}

var sharedKey = []byte("this_is_a_32_byte_example_key!32")

func init() {
	jwtKey := os.Getenv("JWT_KEY")
	if len(jwtKey) == 0 {
		log.Printf("Warning: JWT_KEY environment variable not set, using default value: %s\n", string(sharedKey))
	} else {
		sharedKey = []byte(jwtKey)
	}
}

func CreateJWE(claims TokenClaims) (string, error) {
	recipient := jose.Recipient{
		Algorithm: jose.DIRECT,
		Key:       sharedKey,
	}
	encrypter, err := jose.NewEncrypter(jose.A256GCM, recipient, nil)
	if err != nil {
		return "", err
	}
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	object, err := encrypter.Encrypt(claimsBytes)
	if err != nil {
		return "", err
	}
	jweToken := object.FullSerialize()
	base64Token := base64.StdEncoding.EncodeToString([]byte(jweToken))
	return base64Token, nil
}

func ParseAndValidateJWE(base64Token string) (*TokenClaims, error) {
	jweBytes, err := base64.StdEncoding.DecodeString(base64Token)
	if err != nil {
		return nil, errors.New("error decoding base64 token")
	}
	jweObject, err := jose.ParseEncrypted(string(jweBytes))
	if err != nil {
		return nil, errors.New("invalid token: " + string(jweBytes))
	}
	decryptedBytes, err := jweObject.Decrypt(sharedKey)
	if err != nil {
		return nil, errors.New("unable to decrypt token")
	}

	var claims TokenClaims
	if err := json.Unmarshal(decryptedBytes, &claims); err != nil {
		return nil, errors.New("unable to unmarshal token claims")
	}
	return &claims, nil
}
