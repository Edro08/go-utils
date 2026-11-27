package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func Encrypt(publicKey *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	if publicKey == nil {
		return nil, fmt.Errorf("%w: public key is nil", ErrEncryptFailed)
	}
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, plaintext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptFailed, err)
	}
	return ciphertext, nil
}

func Decrypt(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("%w: private key is nil", ErrDecryptFailed)
	}
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptFailed, err)
	}
	return plaintext, nil
}

func ParsePrivateKeyFromPEM(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, ErrDecodePEMBlockPrivate
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParsePrivateKey, err)
	}
	return key, nil
}

func ParsePublicKeyFromPEM(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, ErrDecodePEMBlockPublic
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParsePublicKey, err)
	}
	publicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("%w: not an RSA public key", ErrParsePublicKey)
	}
	return publicKey, nil
}

func ParsePrivateKeyFromPEMFile(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPrivateKeyNotFound, err)
	}
	return ParsePrivateKeyFromPEM(data)
}

func ParsePublicKeyFromPEMFile(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPublicKeyNotFound, err)
	}
	return ParsePublicKeyFromPEM(data)
}
