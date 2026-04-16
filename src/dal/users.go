package dal

import (
	"flagpole/src/database"
	"flagpole/src/models"
)

type userDAL struct{}

var User = userDAL{}

func (userDAL) Create(email, passwordHash string) error {
	return database.DB.Create(&models.User{Email: email, PasswordHash: passwordHash}).Error
}

func (userDAL) FindByEmail(email string) (models.User, error) {
	var user models.User
	err := database.DB.Where("email = ?", email).First(&user).Error
	return user, err
}
