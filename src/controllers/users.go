package controllers

import (
	"errors"

	"flagpole/src/dal"

	"golang.org/x/crypto/bcrypt"
)

var ErrEmailAlreadyRegistered = errors.New("email already registered")
var ErrInvalidCredentials = errors.New("invalid credentials")

func RegisterUser(email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := dal.User.Create(email, string(hash)); err != nil {
		return ErrEmailAlreadyRegistered
	}
	return nil
}

func AuthenticateUser(email, password string) error {
	user, err := dal.User.FindByEmail(email)
	if err != nil {
		// run bcrypt anyway to prevent timing attacks
		bcrypt.CompareHashAndPassword([]byte("$2a$10$invalid"), []byte(password))
		return ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}
