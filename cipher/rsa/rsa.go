package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

var (
	ErrDecryptFailed         = errors.New("rsa: decrypt failed")
	ErrEncryptFailed         = errors.New("rsa: encrypt failed")
	ErrPrivateKeyNotFound    = errors.New("rsa: private key not found")
	ErrPublicKeyNotFound     = errors.New("rsa: public key not found")
	ErrDecodePEMBlockPrivate = errors.New("rsa: failed to decode PEM block containing private key")
	ErrDecodePEMBlockPublic  = errors.New("rsa: failed to decode PEM block containing public key")
	ErrParsePrivateKey       = errors.New("rsa: failed to parse private key")
	ErrParsePublicKey        = errors.New("rsa: failed to parse public key")
)

// Encrypt cifra plaintext usando RSA-OAEP con SHA-256.
func Encrypt(publicKey *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	if publicKey == nil {
		return nil, fmt.Errorf("%w: public key is nil", ErrEncryptFailed)
	}

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, plaintext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrEncryptFailed, err)
	}

	return ciphertext, nil
}

// Decrypt descifra ciphertext usando RSA-OAEP con SHA-256.
func Decrypt(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("%w: private key is nil", ErrDecryptFailed)
	}

	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrDecryptFailed, err)
	}

	return plaintext, nil
}

// ParsePrivateKeyFromPEM parsea una llave privada RSA en formato PKCS#1 o PKCS#8.
func ParsePrivateKeyFromPEM(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, ErrDecodePEMBlockPrivate
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrParsePrivateKey, err)
		}

		return key, nil

	case "PRIVATE KEY":
		parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrParsePrivateKey, err)
		}

		key, ok := parsed.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("%w: not an RSA private key", ErrParsePrivateKey)
		}

		return key, nil

	default:
		return nil, ErrDecodePEMBlockPrivate
	}
}

// ParsePublicKeyFromPEM parsea una llave pública RSA en formato PKIX o PKCS#1.
func ParsePublicKeyFromPEM(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, ErrDecodePEMBlockPublic
	}

	switch block.Type {
	case "PUBLIC KEY":
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrParsePublicKey, err)
		}

		publicKey, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("%w: not an RSA public key", ErrParsePublicKey)
		}

		return publicKey, nil

	case "RSA PUBLIC KEY":
		publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrParsePublicKey, err)
		}

		return publicKey, nil

	default:
		return nil, ErrDecodePEMBlockPublic
	}
}

// ParsePrivateKeyFromPEMFile lee un archivo PEM y parsea una llave privada RSA.
func ParsePrivateKeyFromPEMFile(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPrivateKeyNotFound, err)
	}

	return ParsePrivateKeyFromPEM(data)
}

// ParsePublicKeyFromPEMFile lee un archivo PEM y parsea una llave pública RSA.
func ParsePublicKeyFromPEMFile(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPublicKeyNotFound, err)
	}

	return ParsePublicKeyFromPEM(data)
}
