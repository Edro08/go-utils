package jwt

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func Test_resolveSigningMethod(t *testing.T) {
	type args struct {
		alg SigningAlgorithm
	}
	tests := []struct {
		name    string
		args    args
		want    jwt.SigningMethod
		wantErr bool
	}{
		{"RS256", args{alg: RS256}, jwt.SigningMethodRS256, false},
		{"RS384", args{alg: RS384}, jwt.SigningMethodRS384, false},
		{"RS512", args{alg: RS512}, jwt.SigningMethodRS512, false},
		{"HS256", args{alg: HS256}, jwt.SigningMethodHS256, false},
		{"HS384", args{alg: HS384}, jwt.SigningMethodHS384, false},
		{"HS512", args{alg: HS512}, jwt.SigningMethodHS512, false},
		{"Invalid algorithm", args{alg: "ES256"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveSigningMethod(tt.args.alg)
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

func TestNew(t *testing.T) {
	type args struct {
		opts NewTokenOpts
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		targetErr error
	}{
		{
			name: "Valid HS256 token with options and reserved claims/headers filter",
			args: args{
				opts: NewTokenOpts{
					ID:        "token-123",
					Algorithm: HS256,
					Issuer:    "my-auth",
					Subject:   "user-456",
					CustomClaims: map[string]any{
						"role": "admin",
						"iss":  "should-be-ignored", // Reserved claim
					},
					CustomHeaders: map[string]string{
						"custom-header": "custom-val",
						"alg":           "should-be-ignored", // Reserved header
					},
					ExpiredEnabled: true,
					ExpiredTime:    1 * time.Hour,
				},
			},
			wantErr:   false,
			targetErr: nil,
		},
		{
			name: "ExpiredEnabled true with zero duration fails",
			args: args{
				opts: NewTokenOpts{
					Algorithm:      HS256,
					ExpiredEnabled: true,
					ExpiredTime:    0,
				},
			},
			wantErr:   true,
			targetErr: ErrExpirationValue,
		},
		{
			name: "ExpiredEnabled true with negative duration fails",
			args: args{
				opts: NewTokenOpts{
					Algorithm:      HS256,
					ExpiredEnabled: true,
					ExpiredTime:    -10 * time.Minute,
				},
			},
			wantErr:   true,
			targetErr: ErrExpirationValue,
		},
		{
			name: "Unsupported algorithm fails",
			args: args{
				opts: NewTokenOpts{
					Algorithm: "NONE",
				},
			},
			wantErr:   true,
			targetErr: ErrSigningMethod,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Errorf("New() error = %v, targetErr %v", err, tt.targetErr)
			}
			if !tt.wantErr && got == nil {
				t.Error("New() returned nil Token for valid options")
			}
		})
	}
}

func TestToken_Sign(t1 *testing.T) {
	secretKey := []byte("my-super-secret-key-32-bytes!!")

	tokenObj, err := New(NewTokenOpts{
		ID:        "t-1",
		Algorithm: HS256,
		Issuer:    "issuer",
	})
	if err != nil {
		t1.Fatalf("failed to create token: %v", err)
	}

	tests := []struct {
		name    string
		token   *Token
		key     any
		wantErr bool
	}{
		{
			name:    "Successful HMAC signing",
			token:   tokenObj,
			key:     secretKey,
			wantErr: false,
		},
		{
			name:    "Invalid key type for HMAC",
			token:   tokenObj,
			key:     12345, // invalid key type for HS256
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			got, err := tt.token.Sign(tt.key)
			if (err != nil) != tt.wantErr {
				t1.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t1.Errorf("Sign() returned empty token string")
			}
		})
	}
}

func TestParse(t *testing.T) {
	secretKey := []byte("test-secret-key")
	tok, err := New(NewTokenOpts{
		ID:        "test-id",
		Algorithm: HS256,
		Issuer:    "test-iss",
		CustomClaims: map[string]any{
			"username": "alice",
		},
	})
	if err != nil {
		t.Fatalf("failed to setup token: %v", err)
	}
	signedString, err := tok.Sign(secretKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	tests := []struct {
		name        string
		tokenString string
		wantErr     bool
		wantKey     string
		wantVal     any
	}{
		{
			name:        "Valid token parses claims correctly",
			tokenString: signedString,
			wantErr:     false,
			wantKey:     "username",
			wantVal:     "alice",
		},
		{
			name:        "Malformed token fails to parse",
			tokenString: "not-a-valid-jwt-token",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.tokenString)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got[tt.wantKey] != tt.wantVal {
					t.Errorf("Parse() claim %s = %v, want %v", tt.wantKey, got[tt.wantKey], tt.wantVal)
				}
			}
		})
	}
}

func TestVerify(t *testing.T) {
	secretKey := []byte("correct-secret")
	wrongKey := []byte("wrong-secret")

	// 1. Valid HS256 token
	hs256Tok, err := New(NewTokenOpts{
		Algorithm: HS256,
		Issuer:    "valid-issuer",
	})
	if err != nil {
		t.Fatalf("failed to create HS256 token: %v", err)
	}
	validHS256String, err := hs256Tok.Sign(secretKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	// 2. Valid HS384 token
	hs384Tok, err := New(NewTokenOpts{
		Algorithm: HS384,
		Issuer:    "valid-issuer",
	})
	if err != nil {
		t.Fatalf("failed to create HS384 token: %v", err)
	}
	validHS384String, err := hs384Tok.Sign(secretKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	// 3. Expired token (manually built with past exp)
	expiredClaims := jwt.MapClaims{
		"iss": "issuer",
		"exp": time.Now().Add(-1 * time.Hour).Unix(),
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, err := expiredToken.SignedString(secretKey)
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}

	type args struct {
		tokenString string
		expectedAlg SigningAlgorithm
		key         any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Successful verification",
			args: args{
				tokenString: validHS256String,
				expectedAlg: HS256,
				key:         secretKey,
			},
			wantErr: false,
		},
		{
			name: "Verification fails with wrong key",
			args: args{
				tokenString: validHS256String,
				expectedAlg: HS256,
				key:         wrongKey,
			},
			wantErr: true,
		},
		{
			name: "Verification fails with algorithm mismatch",
			args: args{
				tokenString: validHS384String,
				expectedAlg: HS256,
				key:         secretKey,
			},
			wantErr: true,
		},
		{
			name: "Verification fails with expired token",
			args: args{
				tokenString: expiredTokenString,
				expectedAlg: HS256,
				key:         secretKey,
			},
			wantErr: true,
		},
		{
			name: "Verification fails with unknown expected algorithm",
			args: args{
				tokenString: validHS256String,
				expectedAlg: SigningAlgorithm("UNKNOWN"),
				key:         secretKey,
			},
			wantErr: true,
		},
		{
			name: "Verification fails with malformed token",
			args: args{
				tokenString: "invalid.token.structure",
				expectedAlg: HS256,
				key:         secretKey,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Verify(tt.args.tokenString, tt.args.expectedAlg, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
