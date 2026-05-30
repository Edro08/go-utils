package jwt

import "errors"

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
	HeaderAlg      = "alg"
	HeaderTyp      = "typ"
	HeaderTypValue = "JWT"

	Iss = "iss"
	Sub = "sub"
	Exp = "exp"
	Nbf = "nbf"
	Iat = "iat"
	JTI = "jti"
)

var ReservedClaims = map[string]struct{}{
	Iss: {}, Exp: {}, Nbf: {}, Iat: {}, JTI: {},
}

var ReservedHeaders = map[string]struct{}{
	HeaderAlg: {}, HeaderTyp: {},
}

var (
	ErrExpirationValue    = errors.New("the expiration value is invalid")
	ErrSignedToken        = errors.New("failed to sign token")
	ErrSigningMethod      = errors.New("signing method not found")
	ErrPrivateKeyNotFound = errors.New("private key not found")
	ErrPublicKeyNotFound  = errors.New("public key not found")
	ErrParsePrivateKey    = errors.New("failed to parse private key")
	ErrParsePublicKey     = errors.New("failed to parse public key")
)
