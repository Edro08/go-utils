package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrPrivateKeyNotFound = errors.New("jwt: private key not found")
	ErrPublicKeyNotFound  = errors.New("jwt: public key not found")
	ErrParsePrivateKey    = errors.New("jwt: failed to parse private key")
	ErrParsePublicKey     = errors.New("jwt: failed to parse public key")
)

// ParseRSAPrivateKeyFromPEM parsea una llave privada RSA desde bytes en formato PEM.
func ParseRSAPrivateKeyFromPEM(keyData []byte) (*rsa.PrivateKey, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrParsePrivateKey, err)
	}
	return key, nil
}

// ParseRSAPublicKeyFromPEM parsea una llave pública RSA desde bytes en formato PEM.
func ParseRSAPublicKeyFromPEM(keyData []byte) (*rsa.PublicKey, error) {
	key, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrParsePublicKey, err)
	}
	return key, nil
}

// ParseRSAPrivateKeyFromPEMFile lee un archivo PEM y parsea una llave privada RSA.
func ParseRSAPrivateKeyFromPEMFile(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPrivateKeyNotFound, err)
	}
	return ParseRSAPrivateKeyFromPEM(keyData)
}

// ParseRSAPublicKeyFromPEMFile lee un archivo PEM y parsea una llave pública RSA.
func ParseRSAPublicKeyFromPEMFile(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPublicKeyNotFound, err)
	}
	return ParseRSAPublicKeyFromPEM(keyData)
}
