package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func generateTestRSAFiles(t *testing.T) (privPath string, pubPath string, invalidPath string) {
	t.Helper()
	tempDir := t.TempDir()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}

	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})

	privPath = filepath.Join(tempDir, "private.pem")
	if err := os.WriteFile(privPath, privPEM, 0600); err != nil {
		t.Fatalf("failed to write private pem: %v", err)
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	pubPath = filepath.Join(tempDir, "public.pem")
	if err := os.WriteFile(pubPath, pubPEM, 0600); err != nil {
		t.Fatalf("failed to write public pem: %v", err)
	}

	invalidPath = filepath.Join(tempDir, "invalid.pem")
	if err := os.WriteFile(invalidPath, []byte("invalid-pem-content"), 0600); err != nil {
		t.Fatalf("failed to write invalid pem: %v", err)
	}

	return privPath, pubPath, invalidPath
}

func TestParseRSAPrivateKeyFromPEMFile(t *testing.T) {
	privPath, _, invalidPath := generateTestRSAFiles(t)

	type args struct {
		path string
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		targetErr error
	}{
		{
			name:      "Valid RSA private key PEM file",
			args:      args{path: privPath},
			wantErr:   false,
			targetErr: nil,
		},
		{
			name:      "Non-existing file returns ErrPrivateKeyNotFound",
			args:      args{path: filepath.Join(t.TempDir(), "non_existent.pem")},
			wantErr:   true,
			targetErr: ErrPrivateKeyNotFound,
		},
		{
			name:      "Invalid PEM content returns ErrParsePrivateKey",
			args:      args{path: invalidPath},
			wantErr:   true,
			targetErr: ErrParsePrivateKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRSAPrivateKeyFromPEMFile(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRSAPrivateKeyFromPEMFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Errorf("ParseRSAPrivateKeyFromPEMFile() error = %v, targetErr %v", err, tt.targetErr)
			}
			if !tt.wantErr && got == nil {
				t.Errorf("ParseRSAPrivateKeyFromPEMFile() returned nil key for valid input")
			}
		})
	}
}

func TestParseRSAPublicKeyFromPEMFile(t *testing.T) {
	_, pubPath, invalidPath := generateTestRSAFiles(t)

	type args struct {
		path string
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		targetErr error
	}{
		{
			name:      "Valid RSA public key PEM file",
			args:      args{path: pubPath},
			wantErr:   false,
			targetErr: nil,
		},
		{
			name:      "Non-existing file returns ErrPublicKeyNotFound",
			args:      args{path: filepath.Join(t.TempDir(), "non_existent.pem")},
			wantErr:   true,
			targetErr: ErrPublicKeyNotFound,
		},
		{
			name:      "Invalid PEM content returns ErrParsePublicKey",
			args:      args{path: invalidPath},
			wantErr:   true,
			targetErr: ErrParsePublicKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRSAPublicKeyFromPEMFile(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRSAPublicKeyFromPEMFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Errorf("ParseRSAPublicKeyFromPEMFile() error = %v, targetErr %v", err, tt.targetErr)
			}
			if !tt.wantErr && got == nil {
				t.Errorf("ParseRSAPublicKeyFromPEMFile() returned nil key for valid input")
			}
		})
	}
}
