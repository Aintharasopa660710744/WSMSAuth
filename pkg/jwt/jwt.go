package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token has expired")
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	UserID string    `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

type Manager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewManager(accessSecret, refreshSecret string, accessExpiry, refreshExpiry time.Duration) *Manager {
	return &Manager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateAccessToken creates a short-lived access JWT
func (m *Manager) GenerateAccessToken(userID uuid.UUID, email, role string) (string, error) {
	return m.generate(userID.String(), email, role, AccessToken, m.accessSecret, m.accessExpiry)
}

// GenerateRefreshToken creates a long-lived refresh JWT
func (m *Manager) GenerateRefreshToken(userID uuid.UUID, email, role string) (string, error) {
	return m.generate(userID.String(), email, role, RefreshToken, m.refreshSecret, m.refreshExpiry)
}

func (m *Manager) generate(userID, email, role string, tokenType TokenType, secret []byte, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ValidateAccessToken parses and validates an access token
func (m *Manager) ValidateAccessToken(tokenStr string) (*Claims, error) {
	return m.validate(tokenStr, m.accessSecret, AccessToken)
}

// ValidateRefreshToken parses and validates a refresh token
func (m *Manager) ValidateRefreshToken(tokenStr string) (*Claims, error) {
	return m.validate(tokenStr, m.refreshSecret, RefreshToken)
}

func (m *Manager) validate(tokenStr string, secret []byte, expectedType TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.Type != expectedType {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// AccessTokenExpiry returns the configured access token expiry in seconds
func (m *Manager) AccessTokenExpiry() int {
	return int(m.accessExpiry.Seconds())
}
