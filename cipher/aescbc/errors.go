package aescbc

import "errors"

var (
	ErrDecryptFailed     = errors.New("decrypt failed")
	ErrEncryptFailed     = errors.New("encrypt failed")
	ErrCreateCipherBlock = errors.New("failed to create cipher block")
)
