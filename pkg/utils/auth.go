package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	Issuer   = "life-easy-api"
	Audience = "life-easy-mobile"
	Lifetime = 15 * time.Minute
)

func HashPassword(password string) (string, error) {
	result, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		12,
	)
	return string(result), err
}

func ComparePassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	) == nil
}

func GenerateToken(userID, secret string) (string, error) {
	now := time.Now()

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    Issuer,
		Audience:  jwt.ClaimStrings{Audience},
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(Lifetime)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func ParseToken(raw, secret string) (string, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(
		raw,
		claims,
		func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(Issuer),
		jwt.WithAudience(Audience),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return "", err
	}

	if !token.Valid || claims.Subject == "" {
		return "", jwt.ErrTokenInvalidClaims
	}

	return claims.Subject, nil
}
