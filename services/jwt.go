package services

import (
	"project/models"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var JwtKey = []byte("my_secret_key")

func GenerateToken(user_id int, username string, role string) (string, error) {
	expirationTime := time.Now().Add(10 * time.Minute)
	claims := &models.Claims{
		UserID:   user_id,
		Username: username,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JwtKey)
}
