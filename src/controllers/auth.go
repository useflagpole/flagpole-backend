package controllers

import (
	"time"

	"flagpole/src/config"
	"flagpole/src/models"

	"github.com/golang-jwt/jwt/v5"
)

const TOKEN_EXPIRY = 24 * time.Hour

func GenerateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"userId": user.ID,
		"exp":    time.Now().Add(TOKEN_EXPIRY).Unix(),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.Get().JWTSecret))
}
