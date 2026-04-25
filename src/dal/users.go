package dal

import (
	"flagpole/src/database"
	"flagpole/src/models"

	"github.com/google/uuid"
)

type userDAL struct{}

var User = userDAL{}

func (userDAL) Create(user *models.User) error {
	return database.DB.Create(user).Error
}

func (userDAL) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := database.DB.Preload("Organizations").Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (userDAL) EmailExists(email string) bool {
	var user models.User
	return database.DB.Where("email = ?", email).First(&user).Error == nil
}

func (userDAL) CountOrganizations(userID uuid.UUID) (int64, error) {
	var count int64
	err := database.DB.Model(&models.UserOrganization{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (userDAL) CountOwnedOrganizations(userID uuid.UUID) (int64, error) {
	var count int64
	err := database.DB.Model(&models.Organization{}).Where("owner_id = ?", userID).Count(&count).Error
	return count, err
}
