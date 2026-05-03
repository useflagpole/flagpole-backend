package dal

import (
	"flagpole/src/database"
	"flagpole/src/models"

	"github.com/google/uuid"
)

type organizationDAL struct{}

var Organization = organizationDAL{}

func (organizationDAL) List() ([]models.Organization, error) {
	var orgs []models.Organization
	if err := database.DB.Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

func (organizationDAL) GetByID(id uint) (*models.Organization, error) {
	var org models.Organization
	if err := database.DB.First(&org, id).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

func (organizationDAL) Create(org *models.Organization) error {
	return database.DB.Create(org).Error
}

func (organizationDAL) Save(org *models.Organization) error {
	return database.DB.Save(org).Error
}

func (organizationDAL) Delete(org *models.Organization) error {
	return database.DB.Delete(org).Error
}

func (organizationDAL) ListByUser(userID uuid.UUID) ([]models.Organization, error) {
	var orgs []models.Organization
	err := database.DB.
		Joins("JOIN org.user_organizations uo ON uo.organization_id = organizations.id").
		Where("uo.user_id = ?", userID).
		Find(&orgs).Error
	if err != nil {
		return nil, err
	}
	return orgs, nil
}

func (organizationDAL) SetPlan(orgID uint, plan string) error {
	return database.DB.Model(&models.Organization{}).Where("id = ?", orgID).Update("plan", plan).Error
}

func (organizationDAL) HasPermission(orgID uint, userID uuid.UUID, permCode string) bool {
	var org models.Organization
	if err := database.DB.First(&org, orgID).Error; err == nil && org.OwnerID == userID {
		return true
	}
	var count int64
	database.DB.Raw(`
		SELECT COUNT(*) FROM org.user_organizations uo
		JOIN org.org_role_permissions orp ON orp.org_role_id = uo.org_role_id
		WHERE uo.user_id = ? AND uo.organization_id = ? AND orp.permission_code = ?
	`, userID, orgID, permCode).Scan(&count)
	return count > 0
}

func (organizationDAL) IsMember(orgID uint, userID uuid.UUID) bool {
	var count int64
	database.DB.Model(&models.UserOrganization{}).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Count(&count)
	return count > 0
}

func (organizationDAL) AddUser(orgID uint, userID uuid.UUID, orgRoleID uint) error {
	return database.DB.Create(&models.UserOrganization{
		OrganizationID: orgID,
		UserID:         userID,
		OrgRoleID:      orgRoleID,
	}).Error
}

type OrgMember struct {
	UserID    uuid.UUID `json:"userId"`
	Username  string    `json:"username"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
}

func (organizationDAL) ListMembers(orgID uint) ([]OrgMember, error) {
	members := make([]OrgMember, 0)
	err := database.DB.Raw(`
		SELECT u.id AS user_id, u.username, u.first_name, u.last_name, u.email, r.name AS role
		FROM auth.users u
		JOIN org.user_organizations uo ON uo.user_id = u.id
		JOIN org.org_roles r ON r.id = uo.org_role_id
		WHERE uo.organization_id = ?
		ORDER BY u.username ASC
	`, orgID).Scan(&members).Error
	return members, err
}
