package jwt

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func Test_resolveSigningMethod(t *testing.T) {
	tests := []struct {
		name    string
		alg     SigningAlgorithm
		want    jwt.SigningMethod
		wantErr bool
	}{
		{"RS256", RS256, jwt.SigningMethodRS256, false},
		{"RS384", RS384, jwt.SigningMethodRS384, false},
		{"RS512", RS512, jwt.SigningMethodRS512, false},
		{"PS256", PS256, jwt.SigningMethodPS256, false},
		{"PS384", PS384, jwt.SigningMethodPS384, false},
		{"PS512", PS512, jwt.SigningMethodPS512, false},
		{"HS256", HS256, jwt.SigningMethodHS256, false},
		{"HS384", HS384, jwt.SigningMethodHS384, false},
		{"HS512", HS512, jwt.SigningMethodHS512, false},
		{"ES256", ES256, jwt.SigningMethodES256, false},
		{"ES384", ES384, jwt.SigningMethodES384, false},
		{"ES512", ES512, jwt.SigningMethodES512, false},
		{"EdDSA", EdDSA, jwt.SigningMethodEdDSA, false},
		{"Invalid algorithm", "none", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveSigningMethod(tt.alg)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveSigningMethod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("resolveSigningMethod() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSign(t *testing.T) {
	secretKey := []byte("my-super-secret-key-32-bytes!!")
	var nilRSAPrivate *rsa.PrivateKey

	tests := []struct {
		name      string
		opts      SignOptions
		wantErr   bool
		targetErr error
	}{
		{
			name: "Valid HS256 token with all options and reserved claim/header protection",
			opts: SignOptions{
				ID:        "  token-123  ",
				Algorithm: HS256,
				Key:       secretKey,
				Issuer:    "  my-auth  ",
				Subject:   "  user-456  ",
				Audience:  []string{"  api-a  ", "", "api-b"},
				CustomClaims: map[string]any{
					"role": "admin",
					"iss":  "should-be-ignored",
					"sub":  "should-be-ignored",
					"aud":  "should-be-ignored",
					"exp":  0,
				},
				CustomHeaders: map[string]string{
					"kid": "key-1",
					"alg": "should-be-ignored",
					"typ": "should-be-ignored",
				},
				Expiration: time.Hour,
			},
			wantErr: false,
		},
		{
			name: "Valid token without Expiration omits exp claim",
			opts: SignOptions{
				Algorithm: HS256,
				Key:       secretKey,
				Issuer:    "my-auth",
			},
			wantErr: false,
		},
		{
			name: "Negative Expiration returns ErrExpirationValue",
			opts: SignOptions{
				Algorithm:  HS256,
				Key:        secretKey,
				Expiration: -10 * time.Minute,
			},
			wantErr:   true,
			targetErr: ErrExpirationValue,
		},
		{
			name: "Unsupported algorithm returns ErrSigningMethod",
			opts: SignOptions{
				Algorithm: "NONE",
				Key:       secretKey,
			},
			wantErr:   true,
			targetErr: ErrSigningMethod,
		},
		{
			name: "Nil key returns ErrInvalidKey",
			opts: SignOptions{
				Algorithm: HS256,
				Key:       nil,
			},
			wantErr:   true,
			targetErr: ErrInvalidKey,
		},
		{
			name: "Empty HMAC byte slice returns ErrInvalidKey",
			opts: SignOptions{
				Algorithm: HS256,
				Key:       []byte{},
			},
			wantErr:   true,
			targetErr: ErrInvalidKey,
		},
		{
			name: "Typed nil RSA private key returns ErrInvalidKey",
			opts: SignOptions{
				Algorithm: RS256,
				Key:       nilRSAPrivate,
			},
			wantErr:   true,
			targetErr: ErrInvalidKey,
		},
		{
			name: "Invalid key type for HS256 returns ErrSignToken",
			opts: SignOptions{
				Algorithm: HS256,
				Key:       12345,
			},
			wantErr:   true,
			targetErr: ErrSignToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStr, err := Sign(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Errorf("Sign() error = %v, targetErr %v", err, tt.targetErr)
			}
			if !tt.wantErr && tokenStr == "" {
				t.Error("Sign() returned empty token string for valid options")
			}
		})
	}
}

func TestSign_ReservedClaimsAndTrimming(t *testing.T) {
	secretKey := []byte("trim-and-reserved-secret")

	tokenStr, err := Sign(SignOptions{
		ID:        "  jti-1  ",
		Algorithm: HS256,
		Key:       secretKey,
		Issuer:    "  auth-svc  ",
		Subject:   "  user-99  ",
		Audience:  []string{"  web  ", "   ", "mobile"},
		CustomClaims: map[string]any{
			"role": "editor",
			"iss":  "evil-issuer",
			"sub":  "evil-subject",
			"aud":  "evil-audience",
		},
		CustomHeaders: map[string]string{
			"kid": "k-2026",
			"alg": "none",
		},
		Expiration: time.Hour,
	})
	if err != nil {
		t.Fatalf("Sign() unexpected error: %v", err)
	}

	claims, headers, err := ParseUnverified(tokenStr)
	if err != nil {
		t.Fatalf("ParseUnverified() unexpected error: %v", err)
	}

	if claims["jti"] != "jti-1" {
		t.Errorf("jti = %v, want %q", claims["jti"], "jti-1")
	}
	if claims["iss"] != "auth-svc" {
		t.Errorf("iss = %v, want %q", claims["iss"], "auth-svc")
	}
	if claims["sub"] != "user-99" {
		t.Errorf("sub = %v, want %q", claims["sub"], "user-99")
	}
	if claims["role"] != "editor" {
		t.Errorf("role = %v, want %q", claims["role"], "editor")
	}
	if _, hasExp := claims["exp"]; !hasExp {
		t.Error("expected exp claim to be present when Expiration > 0")
	}

	wantAud := []any{"web", "mobile"}
	if !reflect.DeepEqual(claims["aud"], wantAud) {
		t.Errorf("aud = %v, want %v", claims["aud"], wantAud)
	}

	if headers["kid"] != "k-2026" {
		t.Errorf("header kid = %v, want %q", headers["kid"], "k-2026")
	}
	if headers["alg"] != string(HS256) {
		t.Errorf("header alg = %v, want %q", headers["alg"], HS256)
	}
}

func TestParseUnverified(t *testing.T) {
	secretKey := []byte("test-secret-key")
	signedString, err := Sign(SignOptions{
		ID:        "test-id",
		Algorithm: HS256,
		Key:       secretKey,
		Issuer:    "test-iss",
		CustomHeaders: map[string]string{
			"kid": "key-42",
		},
		CustomClaims: map[string]any{
			"username": "alice",
		},
	})
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	tests := []struct {
		name        string
		tokenString string
		wantErr     bool
		targetErr   error
	}{
		{
			name:        "Valid token extracts claims and headers",
			tokenString: signedString,
			wantErr:     false,
		},
		{
			name:        "Empty token returns ErrEmptyToken",
			tokenString: "   ",
			wantErr:     true,
			targetErr:   ErrEmptyToken,
		},
		{
			name:        "Malformed token fails to parse with ErrParseToken",
			tokenString: "not-a-valid-jwt-token",
			wantErr:     true,
			targetErr:   ErrParseToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, headers, err := ParseUnverified(tt.tokenString)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUnverified() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Errorf("ParseUnverified() error = %v, targetErr %v", err, tt.targetErr)
			}
			if !tt.wantErr {
				if claims["username"] != "alice" {
					t.Errorf("claims[username] = %v, want %q", claims["username"], "alice")
				}
				if headers["kid"] != "key-42" {
					t.Errorf("headers[kid] = %v, want %q", headers["kid"], "key-42")
				}
			}
		})
	}
}

func TestVerify(t *testing.T) {
	secretKey := []byte("correct-secret")
	wrongKey := []byte("wrong-secret")
	var nilRSAPublic *rsa.PublicKey

	validHS256, err := Sign(SignOptions{
		Algorithm: HS256,
		Key:       secretKey,
		Issuer:    "valid-issuer",
		Subject:   "user-1",
		Audience:  []string{"api-orders", "api-billing"},
		CustomClaims: map[string]any{
			"role": "admin",
		},
		Expiration: time.Hour,
	})
	if err != nil {
		t.Fatalf("failed to sign HS256 token: %v", err)
	}

	validHS384, err := Sign(SignOptions{
		Algorithm: HS384,
		Key:       secretKey,
		Issuer:    "valid-issuer",
	})
	if err != nil {
		t.Fatalf("failed to sign HS384 token: %v", err)
	}

	// Token expirado hace 2 segundos para probar fallo normal y tolerancia con Leeway.
	recentlyExpiredClaims := jwt.MapClaims{
		"iss": "valid-issuer",
		"exp": time.Now().Add(-2 * time.Second).Unix(),
	}
	recentlyExpiredToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, recentlyExpiredClaims).SignedString(secretKey)
	if err != nil {
		t.Fatalf("failed to sign recently expired token: %v", err)
	}

	// Par RSA para verificar flujo asimétrico completo.
	rsaPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}
	validRS256, err := Sign(SignOptions{
		Algorithm:  RS256,
		Key:        rsaPriv,
		Issuer:     "rsa-issuer",
		Expiration: time.Hour,
	})
	if err != nil {
		t.Fatalf("failed to sign RS256 token: %v", err)
	}

	tests := []struct {
		name      string
		opts      VerifyOptions
		wantErr   bool
		targetErr error
	}{
		{
			name: "Successful HS256 verification with Issuer, Subject and Audience",
			opts: VerifyOptions{
				Token:     validHS256,
				Algorithm: HS256,
				Key:       secretKey,
				Issuer:    "valid-issuer",
				Subject:   "user-1",
				Audience:  "api-orders",
			},
			wantErr: false,
		},
		{
			name: "Successful RS256 verification with public key",
			opts: VerifyOptions{
				Token:     validRS256,
				Algorithm: RS256,
				Key:       &rsaPriv.PublicKey,
				Issuer:    "rsa-issuer",
			},
			wantErr: false,
		},
		{
			name: "Recently expired token succeeds within Leeway",
			opts: VerifyOptions{
				Token:     recentlyExpiredToken,
				Algorithm: HS256,
				Key:       secretKey,
				Leeway:    10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "Recently expired token fails without Leeway",
			opts: VerifyOptions{
				Token:     recentlyExpiredToken,
				Algorithm: HS256,
				Key:       secretKey,
			},
			wantErr: true,
		},
		{
			name: "Empty token returns ErrEmptyToken",
			opts: VerifyOptions{
				Token:     "",
				Algorithm: HS256,
				Key:       secretKey,
			},
			wantErr:   true,
			targetErr: ErrEmptyToken,
		},
		{
			name: "Nil key returns ErrInvalidKey",
			opts: VerifyOptions{
				Token:     validHS256,
				Algorithm: HS256,
				Key:       nil,
			},
			wantErr:   true,
			targetErr: ErrInvalidKey,
		},
		{
			name: "Typed nil RSA public key returns ErrInvalidKey",
			opts: VerifyOptions{
				Token:     validRS256,
				Algorithm: RS256,
				Key:       nilRSAPublic,
			},
			wantErr:   true,
			targetErr: ErrInvalidKey,
		},
		{
			name: "Negative leeway returns ErrInvalidLeeway",
			opts: VerifyOptions{
				Token:     validHS256,
				Algorithm: HS256,
				Key:       secretKey,
				Leeway:    -time.Second,
			},
			wantErr:   true,
			targetErr: ErrInvalidLeeway,
		},
		{
			name: "Verification fails with wrong key",
			opts: VerifyOptions{
				Token:     validHS256,
				Algorithm: HS256,
				Key:       wrongKey,
			},
			wantErr:   true,
			targetErr: ErrVerifyToken,
		},
		{
			name: "Verification fails with algorithm mismatch",
			opts: VerifyOptions{
				Token:     validHS384,
				Algorithm: HS256,
				Key:       secretKey,
			},
			wantErr: true,
		},
		{
			name: "Verification fails with issuer mismatch",
			opts: VerifyOptions{
				Token:     validHS256,
				Algorithm: HS256,
				Key:       secretKey,
				Issuer:    "other-issuer",
			},
			wantErr: true,
		},
		{
			name: "Verification fails with subject mismatch",
			opts: VerifyOptions{
				Token:     validHS256,
				Algorithm: HS256,
				Key:       secretKey,
				Subject:   "other-user",
			},
			wantErr: true,
		},
		{
			name: "Verification fails with audience mismatch",
			opts: VerifyOptions{
				Token:     validHS256,
				Algorithm: HS256,
				Key:       secretKey,
				Audience:  "other-api",
			},
			wantErr: true,
		},
		{
			name: "Verification fails with unknown algorithm",
			opts: VerifyOptions{
				Token:     validHS256,
				Algorithm: SigningAlgorithm("UNKNOWN"),
				Key:       secretKey,
			},
			wantErr:   true,
			targetErr: ErrSigningMethod,
		},
		{
			name: "Verification fails with malformed token",
			opts: VerifyOptions{
				Token:     "invalid.token.structure",
				Algorithm: HS256,
				Key:       secretKey,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := Verify(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Errorf("Verify() error = %v, targetErr %v", err, tt.targetErr)
			}
			if !tt.wantErr && claims == nil {
				t.Error("Verify() returned nil claims on valid token")
			}
		})
	}
}

func TestSignAndVerify_ModernAlgorithms(t *testing.T) {
	t.Run("ES256 sign and verify", func(t *testing.T) {
		ecPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("failed to generate ecdsa key: %v", err)
		}

		tokenStr, err := Sign(SignOptions{
			Algorithm:  ES256,
			Key:        ecPriv,
			Issuer:     "ecdsa-svc",
			Expiration: time.Hour,
		})
		if err != nil {
			t.Fatalf("Sign(ES256) error = %v", err)
		}

		claims, err := Verify(VerifyOptions{
			Token:     tokenStr,
			Algorithm: ES256,
			Key:       &ecPriv.PublicKey,
			Issuer:    "ecdsa-svc",
		})
		if err != nil {
			t.Fatalf("Verify(ES256) error = %v", err)
		}
		if claims["iss"] != "ecdsa-svc" {
			t.Errorf("claims[iss] = %v, want %q", claims["iss"], "ecdsa-svc")
		}
	})

	t.Run("EdDSA sign and verify", func(t *testing.T) {
		pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			t.Fatalf("failed to generate ed25519 key: %v", err)
		}

		tokenStr, err := Sign(SignOptions{
			Algorithm:  EdDSA,
			Key:        privKey,
			Issuer:     "ed25519-svc",
			Expiration: time.Hour,
		})
		if err != nil {
			t.Fatalf("Sign(EdDSA) error = %v", err)
		}

		claims, err := Verify(VerifyOptions{
			Token:     tokenStr,
			Algorithm: EdDSA,
			Key:       pubKey,
			Issuer:    "ed25519-svc",
		})
		if err != nil {
			t.Fatalf("Verify(EdDSA) error = %v", err)
		}
		if claims["iss"] != "ed25519-svc" {
			t.Errorf("claims[iss] = %v, want %q", claims["iss"], "ed25519-svc")
		}
	})

	t.Run("PS256 sign and verify", func(t *testing.T) {
		rsaPriv, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatalf("failed to generate rsa key: %v", err)
		}

		tokenStr, err := Sign(SignOptions{
			Algorithm:  PS256,
			Key:        rsaPriv,
			Issuer:     "pss-svc",
			Expiration: time.Hour,
		})
		if err != nil {
			t.Fatalf("Sign(PS256) error = %v", err)
		}

		claims, err := Verify(VerifyOptions{
			Token:     tokenStr,
			Algorithm: PS256,
			Key:       &rsaPriv.PublicKey,
			Issuer:    "pss-svc",
		})
		if err != nil {
			t.Fatalf("Verify(PS256) error = %v", err)
		}
		if claims["iss"] != "pss-svc" {
			t.Errorf("claims[iss] = %v, want %q", claims["iss"], "pss-svc")
		}
	})

	t.Run("Nil or empty ECDSA and Ed25519 keys return ErrInvalidKey", func(t *testing.T) {
		var nilECPriv *ecdsa.PrivateKey
		var nilECPub *ecdsa.PublicKey

		if err := validateKey(nilECPriv); !errors.Is(err, ErrInvalidKey) {
			t.Errorf("validateKey(nilECPriv) = %v, want %v", err, ErrInvalidKey)
		}
		if err := validateKey(nilECPub); !errors.Is(err, ErrInvalidKey) {
			t.Errorf("validateKey(nilECPub) = %v, want %v", err, ErrInvalidKey)
		}
		if err := validateKey(ed25519.PrivateKey{}); !errors.Is(err, ErrInvalidKey) {
			t.Errorf("validateKey(empty ed25519.PrivateKey) = %v, want %v", err, ErrInvalidKey)
		}
		if err := validateKey(ed25519.PublicKey{}); !errors.Is(err, ErrInvalidKey) {
			t.Errorf("validateKey(empty ed25519.PublicKey) = %v, want %v", err, ErrInvalidKey)
		}
	})
}
