package controllers

import (
	"encoding/json"
	"errors"
	"log"

	"flagpole/src/dal"
	"flagpole/src/models"
)

var ErrNotOrgMember  = errors.New("not a member of this organization")
var ErrEmptyName     = errors.New("name cannot be empty")

var DEFAULT_ENVIRONMENTS = []string{"production", "staging", "dev"}

func CreateProject(name string, orgID uint, environments []string) (*models.Project, error) {
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
