/*
Package jwt provee funcionalidades stateless para crear, firmar, analizar y verificar
JSON Web Tokens (JWT) en aplicaciones escritas en Go.

Uso básico:

	tokenString, err := jwt.Sign(jwt.SignOptions{
		ID:         "123",
		Algorithm:  jwt.HS256,
		Key:        []byte("my-secret"),
		Issuer:     "my-service",
		Subject:    "user-42",
		Expiration: time.Hour,
	})
	if err != nil {
		log.Fatal(err)
	}

	claims, err := jwt.Verify(jwt.VerifyOptions{
		Token:     tokenString,
		Algorithm: jwt.HS256,
		Key:       []byte("my-secret"),
		Issuer:    "my-service",
	})
	if err != nil {
		log.Fatal(err)
	}

	rawClaims, headers, err := jwt.ParseUnverified(tokenString)
	if err != nil {
		log.Fatal(err)
	}

Comportamiento:
  - Stateless: la clave y el algoritmo se indican en cada operación, ideal para rotación de llaves (kid).
  - Genera automáticamente los claims iat y nbf.
  - Configura exp mediante Expiration (> 0 expira, 0 sin expiración, < 0 error).
  - Protege los claims y headers reservados para evitar sobrescrituras.
  - Verify valida firma, algoritmo esperado, exp/nbf y opcionalmente iss, sub, aud y leeway, devolviendo los claims en una sola pasada.
  - ParseUnverified extrae claims y headers sin verificar la firma (útil para inspeccionar kid antes de verificar).
*/
package jwt

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type SigningAlgorithm string

const (
	RS256 SigningAlgorithm = "RS256"
	RS384 SigningAlgorithm = "RS384"
	RS512 SigningAlgorithm = "RS512"
	PS256 SigningAlgorithm = "PS256"
	PS384 SigningAlgorithm = "PS384"
	PS512 SigningAlgorithm = "PS512"
	HS256 SigningAlgorithm = "HS256"
	HS384 SigningAlgorithm = "HS384"
	HS512 SigningAlgorithm = "HS512"
	ES256 SigningAlgorithm = "ES256"
	ES384 SigningAlgorithm = "ES384"
	ES512 SigningAlgorithm = "ES512"
	EdDSA SigningAlgorithm = "EdDSA"
)

const (
	headerAlg = "alg"
	headerTyp = "typ"

	iss = "iss"
	sub = "sub"
	aud = "aud"
	exp = "exp"
	nbf = "nbf"
	iat = "iat"
	jti = "jti"
)

var reservedClaims = map[string]struct{}{
	iss: {},
	sub: {},
	aud: {},
	exp: {},
	nbf: {},
	iat: {},
	jti: {},
}

var reservedHeaders = map[string]struct{}{
	headerAlg: {},
	headerTyp: {},
}

var (
	ErrExpirationValue         = errors.New("jwt: expiration value cannot be negative")
	ErrSigningMethod           = errors.New("jwt: signing method not supported")
	ErrUnexpectedSigningMethod = errors.New("jwt: unexpected signing method")
	ErrEmptyToken              = errors.New("jwt: token cannot be empty")
	ErrInvalidKey              = errors.New("jwt: key is nil or empty")
	ErrInvalidLeeway           = errors.New("jwt: leeway cannot be negative")
	ErrSignToken               = errors.New("jwt: failed to sign token")
	ErrVerifyToken             = errors.New("jwt: token verification failed")
	ErrParseToken              = errors.New("jwt: failed to parse unverified token")
)

type SignOptions struct {
	ID            string
	Algorithm     SigningAlgorithm
	Key           any
	Issuer        string
	Subject       string
	Audience      []string
	CustomHeaders map[string]string
	CustomClaims  map[string]any
	Expiration    time.Duration
}

type VerifyOptions struct {
	Token     string
	Algorithm SigningAlgorithm
	Key       any
	Issuer    string
	Subject   string
	Audience  string
	Leeway    time.Duration
}

// Sign crea y firma un token JWT con las opciones proporcionadas.
func Sign(opts SignOptions) (string, error) {
	signingMethod, err := resolveSigningMethod(opts.Algorithm)
	if err != nil {
		return "", err
	}

	if err := validateKey(opts.Key); err != nil {
		return "", err
	}

	if opts.Expiration < 0 {
		return "", ErrExpirationValue
	}

	claims := buildClaims(opts, time.Now())
	token := jwt.NewWithClaims(signingMethod, claims)

	for key, value := range opts.CustomHeaders {
		if _, reserved := reservedHeaders[key]; reserved {
			continue
		}
		token.Header[key] = value
	}

	signed, err := token.SignedString(opts.Key)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrSignToken, err)
	}

	return signed, nil
}

// Verify comprueba la firma, el algoritmo esperado, exp/nbf y las validaciones
// opcionales (Issuer, Subject, Audience, Leeway), devolviendo los claims verificados.
func Verify(opts VerifyOptions) (map[string]any, error) {
	if strings.TrimSpace(opts.Token) == "" {
		return nil, ErrEmptyToken
	}

	method, err := resolveSigningMethod(opts.Algorithm)
	if err != nil {
		return nil, err
	}

	if err := validateKey(opts.Key); err != nil {
		return nil, err
	}

	if opts.Leeway < 0 {
		return nil, ErrInvalidLeeway
	}

	parserOpts := buildParserOptions(method, opts)
	claims := jwt.MapClaims{}

	_, err = jwt.ParseWithClaims(opts.Token, &claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != method.Alg() {
			return nil, fmt.Errorf("%w: got %v, expected %v", ErrUnexpectedSigningMethod, token.Header[headerAlg], opts.Algorithm)
		}
		return opts.Key, nil
	}, parserOpts...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrVerifyToken, err)
	}

	return claims, nil
}

// ParseUnverified extrae los claims y headers de un token sin verificar su firma.
// Útil para inspeccionar headers como "kid" antes de seleccionar la llave para Verify.
func ParseUnverified(tokenString string) (map[string]any, map[string]any, error) {
	if strings.TrimSpace(tokenString) == "" {
		return nil, nil, ErrEmptyToken
	}

	claims := jwt.MapClaims{}
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &claims)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrParseToken, err)
	}

	return claims, token.Header, nil
}

func buildClaims(opts SignOptions, now time.Time) jwt.MapClaims {
	claims := jwt.MapClaims{
		iat: now.Unix(),
		nbf: now.Unix(),
	}

	if issuer := strings.TrimSpace(opts.Issuer); issuer != "" {
		claims[iss] = issuer
	}

	if subject := strings.TrimSpace(opts.Subject); subject != "" {
		claims[sub] = subject
	}

	if id := strings.TrimSpace(opts.ID); id != "" {
		claims[jti] = id
	}

	if audience := normalizeAudience(opts.Audience); len(audience) > 0 {
		claims[aud] = audience
	}

	if opts.Expiration > 0 {
		claims[exp] = now.Add(opts.Expiration).Unix()
	}

	for key, value := range opts.CustomClaims {
		if _, reserved := reservedClaims[key]; reserved {
			continue
		}
		claims[key] = value
	}

	return claims
}

func buildParserOptions(method jwt.SigningMethod, opts VerifyOptions) []jwt.ParserOption {
	parserOpts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{method.Alg()}),
	}

	if issuer := strings.TrimSpace(opts.Issuer); issuer != "" {
		parserOpts = append(parserOpts, jwt.WithIssuer(issuer))
	}

	if subject := strings.TrimSpace(opts.Subject); subject != "" {
		parserOpts = append(parserOpts, jwt.WithSubject(subject))
	}

	if audience := strings.TrimSpace(opts.Audience); audience != "" {
		parserOpts = append(parserOpts, jwt.WithAudience(audience))
	}

	if opts.Leeway > 0 {
		parserOpts = append(parserOpts, jwt.WithLeeway(opts.Leeway))
	}

	return parserOpts
}

func normalizeAudience(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func validateKey(key any) error {
	if key == nil {
		return ErrInvalidKey
	}

	switch typedKey := key.(type) {
	case []byte:
		if len(typedKey) == 0 {
			return ErrInvalidKey
		}
	case *rsa.PrivateKey:
		if typedKey == nil {
			return ErrInvalidKey
		}
	case *rsa.PublicKey:
		if typedKey == nil {
			return ErrInvalidKey
		}
	case *ecdsa.PrivateKey:
		if typedKey == nil {
			return ErrInvalidKey
		}
	case *ecdsa.PublicKey:
		if typedKey == nil {
			return ErrInvalidKey
		}
	case ed25519.PrivateKey:
		if len(typedKey) == 0 {
			return ErrInvalidKey
		}
	case ed25519.PublicKey:
		if len(typedKey) == 0 {
			return ErrInvalidKey
		}
	}

	return nil
}

func resolveSigningMethod(alg SigningAlgorithm) (jwt.SigningMethod, error) {
	switch alg {
	case RS256:
		return jwt.SigningMethodRS256, nil
	case RS384:
		return jwt.SigningMethodRS384, nil
	case RS512:
		return jwt.SigningMethodRS512, nil
	case PS256:
		return jwt.SigningMethodPS256, nil
	case PS384:
		return jwt.SigningMethodPS384, nil
	case PS512:
		return jwt.SigningMethodPS512, nil
	case HS256:
		return jwt.SigningMethodHS256, nil
	case HS384:
		return jwt.SigningMethodHS384, nil
	case HS512:
		return jwt.SigningMethodHS512, nil
	case ES256:
		return jwt.SigningMethodES256, nil
	case ES384:
		return jwt.SigningMethodES384, nil
	case ES512:
		return jwt.SigningMethodES512, nil
	case EdDSA:
		return jwt.SigningMethodEdDSA, nil
	default:
		return nil, ErrSigningMethod
	}
}
