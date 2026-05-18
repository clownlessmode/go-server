package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	AccessTokenTTL      = 15 * time.Minute
	AdminAccessTokenTTL = 3 * time.Hour
	RefreshTokenTTL     = 7 * 24 * time.Hour
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenType    = errors.New("invalid token type")
)

type Claims struct {
	UserID int64  `json:"userId"`
	Login  string `json:"login"`
	Role   string `json:"role"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret []byte
	now    func() time.Time
}

func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{
		secret: []byte(secret),
		now:    time.Now,
	}
}

func (m *TokenManager) GenerateAccessToken(userID int64, login string, role string) (string, time.Time, error) {
	ttl := AccessTokenTTL
	if role == "admin" {
		ttl = AdminAccessTokenTTL
	}

	return m.generateToken(userID, login, role, TokenTypeAccess, ttl)
}

func (m *TokenManager) GenerateRefreshToken(userID int64, login string, role string) (string, time.Time, error) {
	return m.generateToken(userID, login, role, TokenTypeRefresh, RefreshTokenTTL)
}

func (m *TokenManager) ParseAccessToken(tokenString string) (*Claims, error) {
	return m.parseToken(tokenString, TokenTypeAccess)
}

func (m *TokenManager) ParseRefreshToken(tokenString string) (*Claims, error) {
	return m.parseToken(tokenString, TokenTypeRefresh)
}

func (m *TokenManager) generateToken(userID int64, login string, role string, tokenType string, ttl time.Duration) (string, time.Time, error) {
	now := m.now().UTC()
	expiresAt := now.Add(ttl)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
		Login:  login,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	})

	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

func (m *TokenManager) parseToken(tokenString string, expectedType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}

		return m.secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.Type != expectedType {
		return nil, ErrTokenType
	}

	return claims, nil
}
