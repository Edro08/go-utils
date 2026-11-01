package aesgcm

import (
	"errors"
	"reflect"
	"testing"
)

func TestEncrypt(t *testing.T) {
	validKey := []byte("0123456789abcdef0123456789abcdef") // 32 bytes
	invalidKey := []byte("too-short")

	tests := []struct {
		name      string
		key       []byte
		plaintext []byte
		wantErr   bool
		targetErr error
	}{
		{
			name:      "Valid encryption",
			key:       validKey,
			plaintext: []byte("sensitive-data-for-gcm"),
			wantErr:   false,
		},
		{
			name:      "Invalid key size fails with ErrCreateCipherBlock",
			key:       invalidKey,
			plaintext: []byte("data"),
			wantErr:   true,
			targetErr: ErrCreateCipherBlock,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encrypt(tt.key, tt.plaintext)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Errorf("Encrypt() error = %v, targetErr %v", err, tt.targetErr)
			}
			if !tt.wantErr && len(got) == 0 {
				t.Errorf("Encrypt() returned empty ciphertext")
			}
		})
	}
}

func TestDecrypt(t *testing.T) {
	validKey := []byte("0123456789abcdef0123456789abcdef")
	plaintext := []byte("sensitive-data-for-gcm")

	ciphertext, err := Encrypt(validKey, plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt test payload: %v", err)
	}

	corruptedCiphertext := make([]byte, len(ciphertext))
	copy(corruptedCiphertext, ciphertext)
	corruptedCiphertext[len(corruptedCiphertext)-1] ^= 0xFF

	tests := []struct {
		name       string
		key        []byte
		ciphertext []byte
		want       []byte
		wantErr    bool
		targetErr  error
	}{
		{
			name:       "Successful decryption",
			key:        validKey,
			ciphertext: ciphertext,
			want:       plaintext,
			wantErr:    false,
		},
		{
			name:       "Invalid key size fails with ErrCreateCipherBlock",
			key:        []byte("short"),
			ciphertext: ciphertext,
			wantErr:    true,
			targetErr:  ErrCreateCipherBlock,
		},
		{
			name:       "Ciphertext too short fails with ErrDecryptFailed",
			key:        validKey,
			ciphertext: []byte("short"), // less than 12 bytes nonce
			wantErr:    true,
			targetErr:  ErrDecryptFailed,
		},
		{
			name:       "Corrupted ciphertext fails authentication with ErrDecryptFailed",
			key:        validKey,
			ciphertext: corruptedCiphertext,
			wantErr:    true,
			targetErr:  ErrDecryptFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decrypt(tt.key, tt.ciphertext)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Errorf("Decrypt() error = %v, targetErr %v", err, tt.targetErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decrypt() got = %s, want %s", got, tt.want)
			}
		})
	}
}
