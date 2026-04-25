package controllers

import (
	"encoding/json"
	"errors"
	"log"

	"flagpole/src/dal"
	"flagpole/src/models"
)

var ErrNotOrgMember = errors.New("not a member of this organization")

var defaultEnvironments = []string{"production", "staging", "dev"}

func CreateProject(name string, orgID uint) (*models.Project, error) {
	envJSON, err := json.Marshal(defaultEnvironments)
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
