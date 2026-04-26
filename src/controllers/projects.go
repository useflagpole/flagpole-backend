package controllers

import (
	"encoding/json"
	"errors"
	"log"

	"flagpole/src/dal"
	"flagpole/src/models"
)

const MAX_PROJECTS_PER_ORG = 2

var ErrNotOrgMember      = errors.New("not a member of this organization")
var ErrEmptyName         = errors.New("name cannot be empty")
var ErrProjectLimitReached = errors.New("organization has reached the maximum of 2 projects")

var DEFAULT_ENVIRONMENTS = []string{"production", "staging", "dev"}

func CreateProject(name string, orgID uint, environments []string) (*models.Project, error) {
	count, err := dal.Project.CountByOrg(orgID)
	if err != nil {
		return nil, err
	}
	if count >= MAX_PROJECTS_PER_ORG {
		return nil, ErrProjectLimitReached
	}

	envs := environments
	if len(envs) == 0 {
		envs = DEFAULT_ENVIRONMENTS
	}
	envJSON, err := json.Marshal(envs)
	if err != nil {
		log.Printf("CreateProject: marshal envs failed: %v", err)
		return nil, errors.New("internal error")
	}

	p := &models.Project{
		Name:           name,
		OrganizationID: orgID,
		Environments:   string(envJSON),
	}

	if err := dal.Project.Create(p); err != nil {
		return nil, err
	}

	return p, nil
}

func ArchiveProject(proj *models.Project) (*models.Project, error) {
	proj.IsActive = false
	if err := dal.Project.Save(proj); err != nil {
		return nil, err
	}
	return proj, nil
}

func UnarchiveProject(proj *models.Project) (*models.Project, error) {
	count, err := dal.Project.CountByOrg(proj.OrganizationID)
	if err != nil {
		return nil, err
	}
	if count >= MAX_PROJECTS_PER_ORG {
		return nil, ErrProjectLimitReached
	}
	proj.IsActive = true
	if err := dal.Project.Save(proj); err != nil {
		return nil, err
	}
	return proj, nil
}

func RenameProject(proj *models.Project, name string) (*models.Project, error) {
	if name == "" {
		return nil, ErrEmptyName
	}
	proj.Name = name
	if err := dal.Project.Save(proj); err != nil {
		return nil, err
	}
	return proj, nil
}
