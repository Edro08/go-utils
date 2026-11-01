package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrPrivateKeyNotFound = errors.New("private key not found")
	ErrPublicKeyNotFound  = errors.New("public key not found")
	ErrParsePrivateKey    = errors.New("failed to parse private key")
	ErrParsePublicKey     = errors.New("failed to parse public key")
)

func ParseRSAPrivateKeyFromPEMFile(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPrivateKeyNotFound, err)
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParsePrivateKey, err)
	}
	return key, nil
}

func ParseRSAPublicKeyFromPEMFile(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPublicKeyNotFound, err)
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParsePublicKey, err)
	}
	return key, nil
}
