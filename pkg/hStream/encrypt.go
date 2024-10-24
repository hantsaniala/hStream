package hStream

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

// Read RSA 256 public key from function params.
func GetPublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
	pubKeyBlock, rest := pem.Decode([]byte(publicKeyPEM))
	if pubKeyBlock == nil {
		return nil, errors.New("pem.Decode failed: " + string(rest))
	}

	publicKey, err := x509.ParsePKIXPublicKey(pubKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	if publicKey == nil {
		return nil, errors.New("x509.ParsePKIXPublicKey failed")
	}

	return publicKey.(*rsa.PublicKey), nil
}

// Encrypt file using the publicKey from GetPublicKey
func EncryptFile(pubKey *rsa.PublicKey, filepath string) ([]byte, error) {
	if pubKey == nil {
		return []byte{}, errors.New("pubKey cannot be nil")
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return []byte{}, err
	}

	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, data)
	if err != nil {
		return []byte{}, err
	}

	return encrypted, nil
}
