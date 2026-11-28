package aesgcm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

var (
	ErrDecryptFailed     = errors.New("aesgcm: decrypt failed")
	ErrCreateCipherBlock = errors.New("aesgcm: failed to create cipher block")
	ErrCreateGCM         = errors.New("aesgcm: failed to create GCM")
	ErrGenerateNonce     = errors.New("aesgcm: failed to generate nonce")
)

// Encrypt cifra plaintext usando AES-GCM con un nonce aleatorio antepuesto al resultado.
func Encrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCreateCipherBlock, err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCreateGCM, err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrGenerateNonce, err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

// Decrypt extrae el nonce antepuesto y descifra ciphertext usando AES-GCM.
func Decrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCreateCipherBlock, err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCreateGCM, err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("%w: ciphertext too short", ErrDecryptFailed)
	}

	nonce := ciphertext[:nonceSize]
	cipherData := ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrDecryptFailed, err)
	}

	return plaintext, nil
}
