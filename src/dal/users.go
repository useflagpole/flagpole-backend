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

func (userDAL) UsernameExists(username string) bool {
	var user models.User
	return database.DB.Where("username = ?", username).First(&user).Error == nil
}

func (userDAL) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (userDAL) UpdateUsername(id uuid.UUID, username string) error {
	return database.DB.Model(&models.User{}).Where("id = ?", id).Update("username", username).Error
}

func (userDAL) GetOrgRoles(userID uuid.UUID) (map[uint]string, error) {
	type row struct {
		OrganizationID uint
		RoleName       string
	}
	var rows []row
	err := database.DB.
		Table("auth.user_organizations uo").
		Select("uo.organization_id, r.name as role_name").
		Joins("JOIN auth.roles r ON r.id = uo.role_id").
		Where("uo.user_id = ?", userID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]string, len(rows))
	for _, r := range rows {
		result[r.OrganizationID] = r.RoleName
	}
	return result, nil
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
