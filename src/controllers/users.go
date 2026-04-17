package controllers

import (
	"errors"
	"log"

	"flagpole/src/dal"
	"flagpole/src/models"
	"flagpole/src/pkg/crypto"
)

var ErrEmailAlreadyRegistered = errors.New("email already registered")
var ErrInvalidCredentials = errors.New("invalid credentials")

func RegisterUser(email, firstName, lastName, password string, orgID uint) (*models.User, error) {
	if dal.User.EmailExists(email) {
		return nil, ErrEmailAlreadyRegistered
	}

	viewerRole, err := dal.Role.GetByName("viewer")
	if err != nil {
		log.Printf("RegisterUser: viewer role lookup failed: %v", err)
		return nil, errors.New("internal error")
	}

	salt, err := crypto.GenerateSalt()
	if err != nil {
		return nil, err
	}

	hash, err := crypto.HashPassword(password, salt)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		PwdHash:   hash,
		PwdSalt:   salt,
		RoleID:    viewerRole.ID,
		OrgID:     orgID,
	}

	if err := dal.User.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func AuthenticateUser(email, password string) (*models.User, error) {
	user, err := dal.User.GetByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !crypto.VerifyPassword(password, user.PwdSalt, user.PwdHash) {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
