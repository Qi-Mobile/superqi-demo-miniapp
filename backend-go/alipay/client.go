package alipay

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var Interface *Client

type Config struct {
	GatewayURL             string
	MerchantPrivateKeyPath string
	AlipayPublicKeyPath    string
	ClientID               string
}

type Client struct {
	config     Config
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	httpClient *http.Client
}

func loadEnvConfig() (Config, error) {
	gatewayURL := os.Getenv("ALIPAY_GATEWAY_URL")
	if gatewayURL == "" {
		return Config{}, errors.New("ALIPAY_GATEWAY_URL is not set")
	}

	merchantPrivateKeyPath := os.Getenv("ALIPAY_MERCHANT_PRIVATE_KEY_PATH")
	if merchantPrivateKeyPath == "" {
		return Config{}, errors.New("ALIPAY_MERCHANT_PRIVATE_KEY_PATH is not set")
	}

	alipayPublicKeyPath := os.Getenv("ALIPAY_PUBLIC_KEY_PATH")
	if alipayPublicKeyPath == "" {
		return Config{}, errors.New("ALIPAY_PUBLIC_KEY_PATH is not set")
	}

	clientID := os.Getenv("ALIPAY_CLIENT_ID")
	if clientID == "" {
		return Config{}, errors.New("ALIPAY_CLIENT_ID is not set")
	}

	return Config{
		GatewayURL:             gatewayURL,
		MerchantPrivateKeyPath: merchantPrivateKeyPath,
		AlipayPublicKeyPath:    alipayPublicKeyPath,
		ClientID:               clientID,
	}, nil
}

func InitAlipayClient() error {
	config, err := loadEnvConfig()
	if err != nil {
		return err
	}

	privateKey, err := loadPrivateKey(config.MerchantPrivateKeyPath)
	if err != nil {
		return err
	}

	publicKey, err := loadPublicKey(config.AlipayPublicKeyPath)
	if err != nil {
		return err
	}

	Interface = &Client{
		config:     config,
		privateKey: privateKey,
		publicKey:  publicKey,
		httpClient: &http.Client{
			Timeout: time.Second * 25,
		},
	}
	return nil
}

func (client *Client) buildHeaders(method, path string, params interface{}) (map[string]string, error) {
	currentTimestamp := time.Now().Format("2006-01-02T15:04:05-07:00")
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	signature, err := client.generateSignature(method, path, currentTimestamp, string(paramsJSON))
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"Content-Type": "application/json; charset=UTF-8",
		"Client-Id":    client.config.ClientID,
		"Request-Time": currentTimestamp,
		"Signature":    fmt.Sprintf("algorithm=RSA256, keyVersion=1, signature=%s", signature),
	}, nil
}

func (client *Client) generateSignature(httpMethod, path, requestTime, content string) (string, error) {
	signContent := fmt.Sprintf("%s %s\n%s.%s.%s", httpMethod, path, client.config.ClientID, requestTime, content)
	hash := sha256.Sum256([]byte(signContent))

	signature, err := rsa.SignPKCS1v15(nil, client.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func (client *Client) sendRequest(path, method string, headers map[string]string, params map[string]string) ([]byte, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, client.config.GatewayURL+path, strings.NewReader(string(paramsJSON)))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (client *Client) sendRequestWithInterface(path, method string, headers map[string]string, params interface{}) ([]byte, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, client.config.GatewayURL+path, strings.NewReader(string(paramsJSON)))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
