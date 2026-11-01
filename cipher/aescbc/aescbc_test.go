package aescbc

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func Test_pkcs7Pad(t *testing.T) {
	type args struct {
		data      []byte
		blockSize int
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "Data needing partial padding",
			args: args{
				data:      []byte("hello"), // 5 bytes, blockSize 8 -> 3 bytes of 0x03
				blockSize: 8,
			},
			want: []byte{'h', 'e', 'l', 'l', 'o', 3, 3, 3},
		},
		{
			name: "Data matching blockSize gets full block padding",
			args: args{
				data:      []byte("12345678"), // 8 bytes, blockSize 8 -> 8 bytes of 0x08
				blockSize: 8,
			},
			want: []byte{'1', '2', '3', '4', '5', '6', '7', '8', 8, 8, 8, 8, 8, 8, 8, 8},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pkcs7Pad(tt.args.data, tt.args.blockSize); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pkcs7Pad() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_pkcs7Unpad(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    []byte
		wantErr bool
	}{
		{
			name:    "Valid unpad",
			data:    []byte{'h', 'e', 'l', 'l', 'o', 3, 3, 3},
			want:    []byte("hello"),
			wantErr: false,
		},
		{
			name:    "Empty data fails",
			data:    []byte{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Unpadding larger than length fails",
			data:    []byte{10, 10},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pkcs7Unpad(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("pkcs7Unpad() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pkcs7Unpad() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncrypt(t *testing.T) {
	validKey := []byte("0123456789abcdef0123456789abcdef") // 32 bytes AES-256
	validIV := []byte("1234567890abcdef")                 // 16 bytes IV
	invalidKey := []byte("short-key")

	tests := []struct {
		name      string
		key       []byte
		iv        []byte
		plaintext []byte
		wantErr   bool
		targetErr error
	}{
		{
			name:      "Valid encryption",
			key:       validKey,
			iv:        validIV,
			plaintext: []byte("secret-message"),
			wantErr:   false,
		},
		{
			name:      "Invalid key size fails with ErrCreateCipherBlock",
			key:       invalidKey,
			iv:        validIV,
			plaintext: []byte("secret-message"),
			wantErr:   true,
			targetErr: ErrCreateCipherBlock,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encrypt(tt.key, tt.iv, tt.plaintext)
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
	validIV := []byte("1234567890abcdef")
	plaintext := []byte("secret-payload-to-encrypt")

	ciphertext, err := Encrypt(validKey, validIV, plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	tests := []struct {
		name       string
		key        []byte
		iv         []byte
		ciphertext []byte
		want       []byte
		wantErr    bool
		targetErr  error
	}{
		{
			name:       "Successful decryption",
			key:        validKey,
			iv:         validIV,
			ciphertext: bytes.Clone(ciphertext),
			want:       plaintext,
			wantErr:    false,
		},
		{
			name:       "Invalid key size fails",
			key:        []byte("short-key"),
			iv:         validIV,
			ciphertext: bytes.Clone(ciphertext),
			wantErr:    true,
			targetErr:  ErrCreateCipherBlock,
		},
		{
			name:       "Ciphertext not a multiple of block size fails",
			key:        validKey,
			iv:         validIV,
			ciphertext: []byte("not-multiple-of-16"),
			wantErr:    true,
			targetErr:  ErrDecryptFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decrypt(tt.key, tt.iv, tt.ciphertext)
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
