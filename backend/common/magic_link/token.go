package magic_link

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ActionClaims struct {
	IncidentID string `json:"inc_id"`
	ServiceID  uint64 `json:"svc_id"`
	OnCaller   string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateToken(incidentID string, serviceID uint64, email string, secretKey []byte) (string, error) {
	claims := ActionClaims{
		IncidentID: incidentID,
		ServiceID:  serviceID,
		OnCaller:   email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			Issuer:    "alerting-platform",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func ParseToken(tokenString string, secretKey []byte) (*ActionClaims, error) {
	claims := &ActionClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	return claims, nil
}
