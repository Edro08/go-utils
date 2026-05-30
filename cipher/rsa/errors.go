package rsa

import "errors"

var (
	ErrDecryptFailed         = errors.New("decrypt failed")
	ErrEncryptFailed         = errors.New("encrypt failed")
	ErrPrivateKeyNotFound    = errors.New("private key not found")
	ErrPublicKeyNotFound     = errors.New("public key not found")
	ErrDecodePEMBlockPrivate = errors.New("failed to decode PEM block containing private key")
	ErrDecodePEMBlockPublic  = errors.New("failed to decode PEM block containing public key")
	ErrParsePrivateKey       = errors.New("failed to parse private key")
	ErrParsePublicKey        = errors.New("failed to parse public key")
)
