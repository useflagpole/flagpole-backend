package controllers

import (
	"time"

	"flagpole/src/config"
	"flagpole/src/dal"
	"flagpole/src/models"

	"github.com/golang-jwt/jwt/v5"
)

const TOKEN_EXPIRY = 24 * time.Hour

func GenerateToken(user *models.User) (string, error) {
	role, err := dal.Role.GetByID(user.RoleID)
	if err != nil {
		return "", err
	}

	orgIDs := make([]uint, len(user.Organizations))
	orgNames := make([]string, len(user.Organizations))
	for i, org := range user.Organizations {
		orgIDs[i] = org.ID
		orgNames[i] = org.Name
	}

	claims := jwt.MapClaims{
		"userId":    user.ID,
		"username":  user.Username,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"email":     user.Email,
		"role":      role.Name,
		"orgIds":    orgIDs,
		"orgNames":  orgNames,
		"exp":       time.Now().Add(TOKEN_EXPIRY).Unix(),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.Get().JWTSecret))
}
