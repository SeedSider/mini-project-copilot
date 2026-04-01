package manager

import (
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
)

type JWTManager struct {
	secretKey string
}

type UserClaims struct {
	jwt.StandardClaims
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{secretKey}
}

func (m *JWTManager) Verify(accessToken string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}
			return []byte(m.secretKey), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
