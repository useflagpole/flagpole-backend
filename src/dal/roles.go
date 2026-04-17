package dal

import (
	"flagpole/src/database"
	"flagpole/src/models"
)

type roleDAL struct{}

var Role = roleDAL{}

func (roleDAL) GetByID(id uint) (*models.Role, error) {
	var role models.Role
	if err := database.DB.First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (roleDAL) GetByName(name string) (*models.Role, error) {
	var role models.Role
	if err := database.DB.Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}
