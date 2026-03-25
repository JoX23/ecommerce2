package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims holds the parsed JWT payload fields we care about.
type Claims struct {
	UserID string
	Email  string
}

// TokenService generates and validates HS256 JWTs.
type TokenService struct {
	secret string
	ttl    time.Duration
}

// NewTokenService creates a TokenService with a 24h TTL.
func NewTokenService(secret string) *TokenService {
	return &TokenService{secret: secret, ttl: 24 * time.Hour}
}

// Generate signs a new JWT for the given user.
func (s *TokenService) Generate(userID, email string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(s.ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

// Validate parses and validates a JWT string, returning the embedded claims.
func (s *TokenService) Validate(tokenStr string) (*Claims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.secret), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	mc, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return &Claims{
		UserID: mc["sub"].(string),
		Email:  mc["email"].(string),
	}, nil
}
