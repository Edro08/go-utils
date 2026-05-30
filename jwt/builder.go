package jwt

import (
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type NewTokenOpts struct {
	ID             string
	Algorithm      SigningAlgorithm
	Issuer         string
	Subject        string
	CustomHeaders  map[string]string
	CustomClaims   map[string]interface{}
	ExpiredEnabled bool
	ExpiredTime    time.Duration
}

type Token struct {
	jwt *jwt.Token
	err error
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

func New(opts NewTokenOpts) *Token {
	signingAlg, err := resolveSigningMethod(opts.Algorithm)
	if err != nil {
		return &Token{err: err}
	}

	now := time.Now()

	c := jwt.MapClaims{}
	c[Iat] = now.Unix()
	c[Nbf] = now.Unix()

	if strings.TrimSpace(opts.Issuer) != "" {
		c[Iss] = opts.Issuer
	}
	if strings.TrimSpace(opts.Subject) != "" {
		c[Sub] = opts.Subject
	}
	if strings.TrimSpace(opts.ID) != "" {
		c[JTI] = opts.ID
	}
	if opts.ExpiredEnabled {
		if opts.ExpiredTime <= 0 {
			return &Token{err: ErrExpirationValue}
		}
		c[Exp] = now.Add(opts.ExpiredTime).Unix()
	}
	for key, value := range opts.CustomClaims {
		if _, reserved := ReservedClaims[key]; reserved {
			continue
		}
		c[key] = value
	}

	t := jwt.NewWithClaims(signingAlg, c)
	t.Header[HeaderAlg] = signingAlg.Alg()
	t.Header[HeaderTyp] = HeaderTypValue
	for key, value := range opts.CustomHeaders {
		if _, reserved := ReservedHeaders[key]; reserved {
			continue
		}
		t.Header[key] = value
	}
	return &Token{jwt: t}
}

func (t *Token) Sign(key any) (string, error) {
	if t.err != nil {
		return "", t.err
	}
	token, err := t.jwt.SignedString(key)
	if err != nil {
		return "", ErrSignedToken
	}
	return token, nil
}

func (t *Token) Error() error {
	return t.err
}

func Parse(tokenString string) (map[string]interface{}, error) {
	claims := jwt.MapClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(tokenString, &claims)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}(claims), nil
}

func Verify(tokenString string, key any) error {
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return key, nil
	})
	return err
}
