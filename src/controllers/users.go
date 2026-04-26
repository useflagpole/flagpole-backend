package controllers

import (
	"errors"
	"log"

	"flagpole/src/dal"
	"flagpole/src/models"
	"flagpole/src/pkg/crypto"

	"github.com/google/uuid"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUsernameTaken = errors.New("username already taken")

type RegistrationConflict struct {
	EmailTaken    bool
	UsernameTaken bool
}

func (e *RegistrationConflict) Error() string { return "registration conflict" }

func RegisterUser(email, username, firstName, lastName, password string) (*models.User, error) {
	conflict := &RegistrationConflict{
		EmailTaken:    dal.User.EmailExists(email),
		UsernameTaken: dal.User.UsernameExists(username),
	}
	if conflict.EmailTaken || conflict.UsernameTaken {
		return nil, conflict
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
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
		PwdHash:   hash,
		PwdSalt:   salt,
		RoleID:    viewerRole.ID,
	}

	if err := dal.User.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func UpdateUsername(userID uuid.UUID, username string) error {
	if dal.User.UsernameExists(username) {
		return ErrUsernameTaken
	}
	return dal.User.UpdateUsername(userID, username)
}

func GetUserOrganizations(userID uuid.UUID) ([]models.Organization, error) {
	return dal.Organization.ListByUser(userID)
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
