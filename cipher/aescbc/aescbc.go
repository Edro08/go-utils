package aescbc

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/subtle"
	"errors"
	"fmt"
)

var (
	ErrDecryptFailed     = errors.New("aescbc: decrypt failed")
	ErrEncryptFailed     = errors.New("aescbc: encrypt failed")
	ErrCreateCipherBlock = errors.New("aescbc: failed to create cipher block")
)

// Encrypt cifra plaintext usando AES-CBC con padding PKCS#7.
// El IV debe tener exactamente aes.BlockSize (16) bytes.
func Encrypt(key, iv, plaintext []byte) ([]byte, error) {
	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("%w: invalid IV size", ErrEncryptFailed)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCreateCipherBlock, err)
	}

	padded := pkcs7Pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(padded))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	return ciphertext, nil
}

// Decrypt descifra ciphertext usando AES-CBC y remueve el padding PKCS#7.
// El IV debe tener exactamente aes.BlockSize (16) bytes.
func Decrypt(key, iv, ciphertext []byte) ([]byte, error) {
	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("%w: invalid IV size", ErrDecryptFailed)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCreateCipherBlock, err)
	}

	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("%w: ciphertext is not a valid multiple of the block size", ErrDecryptFailed)
	}

	plaintext := make([]byte, len(ciphertext))

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	return pkcs7Unpad(plaintext)
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padded := make([]byte, len(data)+padding)

	copy(padded, data)
	copy(padded[len(data):], bytes.Repeat([]byte{byte(padding)}, padding))

	return padded
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("%w: invalid padding", ErrDecryptFailed)
	}

	unpadding := int(data[length-1])
	if unpadding == 0 || unpadding > aes.BlockSize || unpadding > length {
		return nil, fmt.Errorf("%w: invalid padding", ErrDecryptFailed)
	}

	var valid byte = 1
	for i := length - unpadding; i < length; i++ {
		valid &= byte(subtle.ConstantTimeByteEq(data[i], byte(unpadding)))
	}

	if valid != 1 {
		return nil, fmt.Errorf("%w: invalid padding", ErrDecryptFailed)
	}

	return data[:length-unpadding], nil
}
