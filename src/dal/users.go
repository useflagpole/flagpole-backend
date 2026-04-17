package dal

import (
	"flagpole/src/database"
	"flagpole/src/models"
)

type userDAL struct{}

var User = userDAL{}

func (userDAL) Create(user *models.User) error {
	return database.DB.Create(user).Error
}

func (userDAL) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (userDAL) EmailExists(email string) bool {
	var user models.User
	return database.DB.Where("email = ?", email).First(&user).Error == nil
}
