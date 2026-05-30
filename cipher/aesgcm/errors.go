package aesgcm

import "errors"

var (
	ErrDecryptFailed     = errors.New("decrypt failed")
	ErrEncryptFailed     = errors.New("encrypt failed")
	ErrCreateCipherBlock = errors.New("failed to create cipher block")
	ErrCreateGCM         = errors.New("failed to create GCM")
	ErrGenerateNonce     = errors.New("failed to generate nonce")
)
