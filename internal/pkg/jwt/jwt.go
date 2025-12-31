package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yourorg/anonymous-support/internal/domain"
)

type JWTManager struct {
	secret        []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	issuer        string
	audience      string
}

// Manager is an alias for JWTManager for backward compatibility
type Manager = JWTManager

func NewJWTManager(secret string, accessExpiry, refreshExpiry time.Duration) *JWTManager {
	return &JWTManager{
		secret:        []byte(secret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
		issuer:        "anonymous-support-api",
		audience:      "anonymous-support-client",
	}
}

// NewManager is an alias for NewJWTManager for backward compatibility
func NewManager(secret string, accessExpiry, refreshExpiry time.Duration) *JWTManager {
	return NewJWTManager(secret, accessExpiry, refreshExpiry)
}

type Claims struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	IsAnonymous bool   `json:"is_anonymous"`
	Role        string `json:"role"`
	jwt.RegisteredClaims
}

func (m *JWTManager) GenerateAccessToken(user *domain.User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:      user.ID.String(),
		Username:    user.Username,
		IsAnonymous: user.IsAnonymous,
		Role:        string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Audience:  jwt.ClaimStrings{m.audience},
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    m.issuer,
		Audience:  jwt.ClaimStrings{m.audience},
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshExpiry)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithAudience(m.audience), jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Additional validation for NotBefore
		if claims.NotBefore != nil && claims.NotBefore.After(time.Now()) {
			return nil, fmt.Errorf("token not yet valid")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (m *JWTManager) ValidateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithAudience(m.audience), jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		// Additional validation for NotBefore
		if claims.NotBefore != nil && claims.NotBefore.After(time.Now()) {
			return "", fmt.Errorf("token not yet valid")
		}
		return claims.Subject, nil
	}

	return "", fmt.Errorf("invalid refresh token")
}
