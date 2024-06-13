package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func GenerateECDSAKey(path string) ([]byte, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {

		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, err
		}

		err = SaveECDSAKey(privateKey, path)
		if err != nil {
			return nil, err
		}
		SecretKey, err := x509.MarshalPKCS8PrivateKey(privateKey)
		return SecretKey, nil
	}

	key, err := LoadECDSAKey(path)
	if err != nil {
		return nil, err
	}
	SecretKey, err := x509.MarshalPKCS8PrivateKey(key)
	return SecretKey, nil
}

func SaveECDSAKey(key *ecdsa.PrivateKey, filename string) error {
	keyBytes, err := x509.MarshalECPrivateKey(key)

	if err != nil {
		return err
	}

	pemBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = pem.Encode(file, pemBlock)
	if err != nil {
		return err
	}

	return nil
}

func LoadECDSAKey(filePath string) (*ecdsa.PrivateKey, error) {
	keyData, err := os.ReadFile(filePath)

	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}
