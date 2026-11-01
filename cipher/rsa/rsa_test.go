package rsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func generateTestKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey, []byte, []byte) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	return priv, &priv.PublicKey, privPEM, pubPEM
}

func TestEncrypt(t *testing.T) {
	_, pub, _, _ := generateTestKeyPair(t)

	tests := []struct {
		name      string
		publicKey *rsa.PublicKey
		plaintext []byte
		wantErr   bool
		targetErr error
	}{
		{
			name:      "Valid encryption",
			publicKey: pub,
			plaintext: []byte("hello-rsa-world"),
			wantErr:   false,
		},
		{
			name:      "Message too long for RSA key size returns ErrEncryptFailed",
			publicKey: pub,
			plaintext: make([]byte, 512), // 512 bytes exceeds 2048-bit PKCS1v15 max message length
			wantErr:   true,
			targetErr: ErrEncryptFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encrypt(tt.publicKey, tt.plaintext)
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
	priv, pub, _, _ := generateTestKeyPair(t)
	plaintext := []byte("secret-payload-for-rsa")

	ciphertext, err := Encrypt(pub, plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt test payload: %v", err)
	}

	tests := []struct {
		name       string
		privateKey *rsa.PrivateKey
		ciphertext []byte
		want       []byte
		wantErr    bool
		targetErr  error
	}{
		{
			name:       "Successful decryption",
			privateKey: priv,
			ciphertext: ciphertext,
			want:       plaintext,
			wantErr:    false,
		},
		{
			name:       "Corrupted ciphertext returns ErrDecryptFailed",
			privateKey: priv,
			ciphertext: []byte("invalid-ciphertext-bytes"),
			wantErr:    true,
			targetErr:  ErrDecryptFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decrypt(tt.privateKey, tt.ciphertext)
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

func TestParsePrivateKeyFromPEM(t *testing.T) {
	_, _, privPEM, pubPEM := generateTestKeyPair(t)

	corruptedPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: []byte("corrupted-der-bytes"),
	})

	tests := []struct {
		name      string
		pemData   []byte
		wantErr   bool
		targetErr error
	}{
		{
			name:      "Valid private key PEM",
			pemData:   privPEM,
			wantErr:   false,
		},
		{
			name:      "Wrong block type returns ErrDecodePEMBlockPrivate",
			pemData:   pubPEM,
			wantErr:   true,
			targetErr: ErrDecodePEMBlockPrivate,
		},
		{
			name:      "Corrupted DER bytes returns ErrParsePrivateKey",
			pemData:   corruptedPEM,
			wantErr:   true,
			targetErr: ErrParsePrivateKey,
		},
		{
			name:      "Non-PEM data returns ErrDecodePEMBlockPrivate",
			pemData:   []byte("not-a-pem"),
			wantErr:   true,
			targetErr: ErrDecodePEMBlockPrivate,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePrivateKeyFromPEM(tt.pemData)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePrivateKeyFromPEM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Errorf("ParsePrivateKeyFromPEM() error = %v, targetErr %v", err, tt.targetErr)
			}
			if !tt.wantErr && got == nil {
				t.Errorf("ParsePrivateKeyFromPEM() returned nil key")
			}
		})
	}
}

func TestParsePublicKeyFromPEM(t *testing.T) {
	_, _, privPEM, pubPEM := generateTestKeyPair(t)

	corruptedPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: []byte("corrupted-der-bytes"),
	})

	ecdsaKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	ecdsaPubBytes, _ := x509.MarshalPKIXPublicKey(&ecdsaKey.PublicKey)
	ecdsaPubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: ecdsaPubBytes,
	})

	tests := []struct {
		name      string
		pemData   []byte
		wantErr   bool
		targetErr error
	}{
		{
			name:      "Valid public key PEM",
			pemData:   pubPEM,
			wantErr:   false,
		},
		{
			name:      "Non-RSA public key returns ErrParsePublicKey",
			pemData:   ecdsaPubPEM,
			wantErr:   true,
			targetErr: ErrParsePublicKey,
		},
		{
			name:      "Wrong block type returns ErrDecodePEMBlockPublic",
			pemData:   privPEM,
			wantErr:   true,
			targetErr: ErrDecodePEMBlockPublic,
		},
		{
			name:      "Corrupted DER bytes returns ErrParsePublicKey",
			pemData:   corruptedPEM,
			wantErr:   true,
			targetErr: ErrParsePublicKey,
		},
		{
			name:      "Non-PEM data returns ErrDecodePEMBlockPublic",
			pemData:   []byte("not-a-pem"),
			wantErr:   true,
			targetErr: ErrDecodePEMBlockPublic,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePublicKeyFromPEM(tt.pemData)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePublicKeyFromPEM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Errorf("ParsePublicKeyFromPEM() error = %v, targetErr %v", err, tt.targetErr)
			}
			if !tt.wantErr && got == nil {
				t.Errorf("ParsePublicKeyFromPEM() returned nil key")
			}
		})
	}
}

func TestParsePrivateKeyFromPEMFile(t *testing.T) {
	tempDir := t.TempDir()
	_, _, privPEM, _ := generateTestKeyPair(t)

	validPath := filepath.Join(tempDir, "private.pem")
	if err := os.WriteFile(validPath, privPEM, 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tests := []struct {
		name      string
		path      string
		wantErr   bool
		targetErr error
	}{
		{
			name:    "Valid file parses key",
			path:    validPath,
			wantErr: false,
		},
		{
			name:      "Non-existent file returns ErrPrivateKeyNotFound",
			path:      filepath.Join(tempDir, "not_found.pem"),
			wantErr:   true,
			targetErr: ErrPrivateKeyNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePrivateKeyFromPEMFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePrivateKeyFromPEMFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Errorf("ParsePrivateKeyFromPEMFile() error = %v, targetErr %v", err, tt.targetErr)
			}
			if !tt.wantErr && got == nil {
				t.Errorf("ParsePrivateKeyFromPEMFile() returned nil key")
			}
		})
	}
}

func TestParsePublicKeyFromPEMFile(t *testing.T) {
	tempDir := t.TempDir()
	_, _, _, pubPEM := generateTestKeyPair(t)

	validPath := filepath.Join(tempDir, "public.pem")
	if err := os.WriteFile(validPath, pubPEM, 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tests := []struct {
		name      string
		path      string
		wantErr   bool
		targetErr error
	}{
		{
			name:    "Valid file parses public key",
			path:    validPath,
			wantErr: false,
		},
		{
			name:      "Non-existent file returns ErrPublicKeyNotFound",
			path:      filepath.Join(tempDir, "not_found.pem"),
			wantErr:   true,
			targetErr: ErrPublicKeyNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePublicKeyFromPEMFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePublicKeyFromPEMFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Errorf("ParsePublicKeyFromPEMFile() error = %v, targetErr %v", err, tt.targetErr)
			}
			if !tt.wantErr && got == nil {
				t.Errorf("ParsePublicKeyFromPEMFile() returned nil key")
			}
		})
	}
}
