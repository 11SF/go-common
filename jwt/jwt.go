package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTOptions struct {
	AccessTokenExpiredTime  time.Duration
	RefreshTokenExpiredTime time.Duration
	AccessTokenSecretKey    []byte
	RefreshTokenSecretKey   []byte
}

type JWT struct {
	AccessTokenExpiredTime  time.Duration
	RefreshTokenExpiredTime time.Duration
	AccessTokenSecretKey    []byte
	RefreshTokenSecretKey   []byte
}

type IJWT interface {
	MakeClaims(userId uuid.UUID) (*CustomClaims, *CustomClaims, uuid.UUID)
	SignAccessTokenWithHS256(claims jwt.Claims) (string, error)
	SignRefreshTokenWithHS256(claims jwt.Claims) (string, error)
	VerifyAccessTokenWithHS256(tokenString string) (*CustomClaims, string, error)
	VerifyRefreshTokenWithHS256(tokenString string) (*CustomClaims, string, error)
	ExtractBearerToken(tokenString string) (string, error)
}

type CustomClaims struct {
	jwt.RegisteredClaims
}

func NewJWT(opts *JWTOptions) IJWT {
	return &JWT{
		AccessTokenExpiredTime:  opts.AccessTokenExpiredTime,
		RefreshTokenExpiredTime: opts.RefreshTokenExpiredTime,
		AccessTokenSecretKey:    opts.AccessTokenSecretKey,
		RefreshTokenSecretKey:   opts.RefreshTokenSecretKey,
	}
}

func (j *JWT) MakeClaims(userId uuid.UUID) (*CustomClaims, *CustomClaims, uuid.UUID) {
	sessionId, err := uuid.NewV7()
	if err != nil {
		sessionId = uuid.New()
	}
	accessTokenClaims := &CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        sessionId.String(),
			Subject:   userId.String(),
			Issuer:    "accounts.ns-auth",
			Audience:  jwt.ClaimStrings{"platform.nspublic.xyz"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.AccessTokenExpiredTime)),
		},
	}
	refreshtokenClaims := &CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        sessionId.String(),
			Subject:   userId.String(),
			Issuer:    "accounts.ns-auth",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.RefreshTokenExpiredTime)),
		},
	}
	return accessTokenClaims, refreshtokenClaims, sessionId
}

func (j *JWT) SignAccessTokenWithHS256(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.AccessTokenSecretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (j *JWT) SignRefreshTokenWithHS256(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.RefreshTokenSecretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (j *JWT) VerifyAccessTokenWithHS256(tokenString string) (*CustomClaims, string, error) {

	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return j.AccessTokenSecretKey, nil
	})

	if err != nil {
		return nil, "", err
	}

	if !token.Valid {
		return nil, "", err
	}

	return claims, token.Raw, nil
}

func (j *JWT) VerifyRefreshTokenWithHS256(tokenString string) (*CustomClaims, string, error) {

	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return j.RefreshTokenSecretKey, nil
	})

	if err != nil {
		return nil, "", err
	}

	if !token.Valid {
		return nil, "", err
	}

	return claims, token.Raw, nil
}

func (j *JWT) ExtractBearerToken(tokenString string) (string, error) {

	token := tokenString[7:]
	if len(token) == 0 {
		return "", errors.New("token is empty")
	}

	return token, nil
}
