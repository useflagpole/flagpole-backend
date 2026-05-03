package controllers

import (
	"errors"

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

type UserOrgDTO struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Role    string `json:"role"`
	IsOwner bool   `json:"isOwner"`
}

type UserDTO struct {
	ID        uuid.UUID    `json:"id"`
	Username  string       `json:"username"`
	FirstName string       `json:"firstName"`
	LastName  string       `json:"lastName"`
	Email     string       `json:"email"`
	Orgs      []UserOrgDTO `json:"orgs"`
}

func GetUser(userID uuid.UUID) (*UserDTO, error) {
	user, err := dal.User.GetByID(userID)
	if err != nil {
		return nil, err
	}

	orgRoleMap, err := dal.User.GetOrgRoles(userID)
	if err != nil {
		return nil, err
	}

	orgs, err := dal.Organization.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	orgDTOs := make([]UserOrgDTO, len(orgs))
	for i, o := range orgs {
		orgDTOs[i] = UserOrgDTO{ID: o.ID, Name: o.Name, Role: orgRoleMap[o.ID], IsOwner: o.OwnerID == userID}
	}

	return &UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Orgs:      orgDTOs,
	}, nil
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
