package auth

import (
	"fmt"
	"strings"
	"time"

	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	Issuer    string `arg:"env:JWT_ISSUER" default:"ihle.cloud"`
	SecretKey string `arg:"env:JWT_SECRET_KEY" `
	Duration  time.Duration
}

func (c JWT) newClaims(account *Account, audience ...string) Claims {

	now := time.Now()

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    c.Issuer,
			Subject:   account.ID,
			Audience:  jwt.ClaimStrings(audience),
			ExpiresAt: jwt.NewNumericDate(now.Add(c.Duration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        "",
		},
		Permissions: map[string]struct{}{},
		ID:          account.ID,
		Name:        "account.Credentials.Username",
		Email:       "account.Email",
	}
	for _, k := range audience {
		claims.Permissions[k] = struct{}{}
	}
	return claims
}

type Claims struct {
	jwt.RegisteredClaims

	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`

	Permissions map[string]struct{} `json:"permissions"`
}

func (c Claims) SignedString(secretKey string) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	return token.SignedString([]byte(secretKey))
}

// func createToken(sub, audience string) *jwt.Token {

// 	now := time.Now()

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
// 		Issuer:    "ihle.cloud",
// 		Subject:   sub,
// 		Audience:  jwt.ClaimStrings{audience},
// 		ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Hour)),
// 		// NotBefore: jwt.NewNumericDate(now),
// 		IssuedAt: jwt.NewNumericDate(now),
// 		ID:       "",
// 	})

// 	return token
// }

// func (a *Authenticator) createJWT(sub, audience string) (string, error) {

// 	now := time.Now()

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
// 		Issuer:    "ihle.cloud",
// 		Subject:   sub,
// 		Audience:  jwt.ClaimStrings{audience},
// 		ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Hour)),
// 		// NotBefore: jwt.NewNumericDate(now),
// 		IssuedAt: jwt.NewNumericDate(now),
// 		ID:       "",
// 	})

// 	return token.SignedString([]byte(jwtconfig.SecretKey))
// }

func (a *JWT) ParseClaims(signedstring string) (*Claims, error) {
	if len(signedstring) > len("Bearer ") && strings.HasPrefix(signedstring, "Bearer ") {
		signedstring = signedstring[len("Bearer "):]
	}

	var claims Claims
	token, err := jwt.ParseWithClaims(signedstring, &claims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.SecretKey), nil
	})

	switch {

	case token == nil || errors.Is(err, jwt.ErrTokenMalformed):
		return nil, fmt.Errorf("that's not even a token: %s", signedstring)

	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		// Invalid signature
		return nil, fmt.Errorf("invalid signature: %s", err.Error())

	case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
		// Token is either expired or not active yet
		return nil, fmt.Errorf("timing is everything: %s", err.Error())

	case err != nil:
		return nil, fmt.Errorf("there's something strange in the neighborhood: %w", err)

	case !token.Valid:
		return nil, fmt.Errorf("invalid %w token", err)
	}

	if _, ok := token.Claims.(*Claims); ok {
		// fmt.Printf("verified claim: %v\n", claims)
	} else {
		return nil, errors.New("invalid claims type")
	}

	return &claims, nil
}
