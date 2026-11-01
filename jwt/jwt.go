/*
Package jwt provee funcionalidades para crear, firmar, analizar y verificar JSON Web Tokens (JWT) en aplicaciones escritas en Go.

Uso básico:

	token, err := jwt.New(jwt.NewTokenOpts{
		ID:             "123",
		Algorithm:      jwt.HS256,
		Issuer:         "my-service",
		Subject:        "user",
		ExpiredEnabled: true,
		ExpiredTime:    time.Hour,
	})

	if err != nil {
		log.Fatal(err)
	}

	tokenString, err := token.Sign([]byte("my-secret"))
	if err != nil {
		log.Fatal(err)
	}

	err = jwt.Verify(tokenString, jwt.HS256, []byte("my-secret"))
	if err != nil {
		log.Fatal(err)
	}

	claims, err := jwt.Parse(tokenString)
	if err != nil {
		log.Fatal(err)
	}

Comportamiento:
  - Permite crear tokens JWT utilizando diferentes algoritmos de firma.
  - Permite configurar issuer, subject, JWT ID y claims personalizados.
  - Permite agregar headers personalizados al token.
  - Genera automáticamente los claims iat y nbf.
  - Permite configurar opcionalmente el claim exp mediante una duración.
  - Protege los claims y headers reservados para evitar que sean sobrescritos.
  - Permite firmar un token utilizando la clave proporcionada.
  - Permite extraer los claims de un token sin verificar su firma mediante Parse.
  - Permite verificar la firma y el algoritmo esperado mediante Verify.
  - Retorna un error cuando el algoritmo solicitado no está soportado,
  - la configuración de expiración es inválida o la verificación del token falla.
*/
package jwt

import (
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
	HS256 SigningAlgorithm = "HS256"
	HS384 SigningAlgorithm = "HS384"
	HS512 SigningAlgorithm = "HS512"
)

const (
	headerAlg = "alg"
	headerTyp = "typ"

	iss = "iss"
	sub = "sub"
	exp = "exp"
	nbf = "nbf"
	iat = "iat"
	jti = "jti"
)

var reservedClaims = map[string]struct{}{
	iss: {}, exp: {}, nbf: {}, iat: {}, jti: {},
}

var reservedHeaders = map[string]struct{}{
	headerAlg: {}, headerTyp: {},
}

var (
	ErrExpirationValue = errors.New("the expiration value is invalid")
	ErrSigningMethod   = errors.New("signing method not found")
)

type NewTokenOpts struct {
	ID             string
	Algorithm      SigningAlgorithm
	Issuer         string
	Subject        string
	CustomHeaders  map[string]string
	CustomClaims   map[string]any
	ExpiredEnabled bool
	ExpiredTime    time.Duration
}

type Token struct {
	inner *jwt.Token
}

// New crea y valida una nueva estructura de Token JWT.
func New(opts NewTokenOpts) (*Token, error) {
	signingAlg, err := resolveSigningMethod(opts.Algorithm)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	c := jwt.MapClaims{
		iat: now.Unix(),
		nbf: now.Unix(),
	}

	if strings.TrimSpace(opts.Issuer) != "" {
		c[iss] = opts.Issuer
	}
	if strings.TrimSpace(opts.Subject) != "" {
		c[sub] = opts.Subject
	}
	if strings.TrimSpace(opts.ID) != "" {
		c[jti] = opts.ID
	}

	if opts.ExpiredEnabled {
		if opts.ExpiredTime <= 0 {
			return nil, ErrExpirationValue
		}
		c[exp] = now.Add(opts.ExpiredTime).Unix()
	}

	for key, value := range opts.CustomClaims {
		if _, reserved := reservedClaims[key]; reserved {
			continue
		}
		c[key] = value
	}

	t := jwt.NewWithClaims(signingAlg, c)

	for key, value := range opts.CustomHeaders {
		if _, reserved := reservedHeaders[key]; reserved {
			continue
		}
		t.Header[key] = value
	}

	return &Token{inner: t}, nil
}

// Sign firma el token con la clave proporcionada.
func (t *Token) Sign(key any) (string, error) {
	token, err := t.inner.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return token, nil
}

// Parse extrae los claims de un token sin verificar su firma.
func Parse(tokenString string) (map[string]any, error) {
	claims := jwt.MapClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(tokenString, &claims)
	if err != nil {
		return nil, fmt.Errorf("failed to parse unverified token: %w", err)
	}
	return claims, nil
}

// Verify comprueba la firma y validez de un token asegurando el tipo de algoritmo esperado.
func Verify(tokenString string, expectedAlg SigningAlgorithm, key any) error {
	method, err := resolveSigningMethod(expectedAlg)
	if err != nil {
		return err
	}

	_, err = jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != method.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v, expected %v", token.Header["alg"], expectedAlg)
		}
		return key, nil
	})
	if err != nil {
		return fmt.Errorf("token verification failed: %w", err)
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
	case HS256:
		return jwt.SigningMethodHS256, nil
	case HS384:
		return jwt.SigningMethodHS384, nil
	case HS512:
		return jwt.SigningMethodHS512, nil
	default:
		return nil, ErrSigningMethod
	}
}
