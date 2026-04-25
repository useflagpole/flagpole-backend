package controllers

import (
	"encoding/json"
	"errors"

	"flagpole/src/dal"
	"flagpole/src/models"
)

const MAX_ENVIRONMENTS = 5

var ErrMaxEnvironments    = errors.New("maximum of 5 environments reached")
var ErrEnvAlreadyExists   = errors.New("environment already exists")
var ErrEnvNotFound        = errors.New("environment not found")
var ErrEnvProtected       = errors.New("production environment is protected")

func parseEnvs(proj *models.Project) ([]string, error) {
	var envs []string
	if err := json.Unmarshal([]byte(proj.Environments), &envs); err != nil {
		return nil, err
	}
	return envs, nil
}

func persistEnvs(proj *models.Project, envs []string) ([]string, error) {
	raw, err := json.Marshal(envs)
	if err != nil {
		return nil, err
	}
	proj.Environments = string(raw)
	if err := dal.Project.Save(proj); err != nil {
		return nil, err
	}
	return envs, nil
}

func ListEnvironments(proj *models.Project) ([]string, error) {
	return parseEnvs(proj)
}

func CreateEnvironment(proj *models.Project, name string) ([]string, error) {
	envs, err := parseEnvs(proj)
	if err != nil {
		return nil, err
	}
	if len(envs) >= MAX_ENVIRONMENTS {
		return nil, ErrMaxEnvironments
	}
	for _, e := range envs {
		if e == name {
			return nil, ErrEnvAlreadyExists
		}
	}
	return persistEnvs(proj, append(envs, name))
}

func RenameEnvironment(proj *models.Project, oldName, newName string) ([]string, error) {
	if oldName == "production" {
		return nil, ErrEnvProtected
	}
	envs, err := parseEnvs(proj)
	if err != nil {
		return nil, err
	}
	for _, e := range envs {
		if e == newName {
			return nil, ErrEnvAlreadyExists
		}
	}
	found := false
	updated := make([]string, len(envs))
	for i, e := range envs {
		if e == oldName {
			updated[i] = newName
			found = true
		} else {
			updated[i] = e
		}
	}
	if !found {
		return nil, ErrEnvNotFound
	}
	return persistEnvs(proj, updated)
}

func DeleteEnvironment(proj *models.Project, name string) ([]string, error) {
	if name == "production" {
		return nil, ErrEnvProtected
	}
	envs, err := parseEnvs(proj)
	if err != nil {
		return nil, err
	}
	found := false
	updated := make([]string, 0, len(envs))
	for _, e := range envs {
		if e == name {
			found = true
		} else {
			updated = append(updated, e)
		}
	}
	if !found {
		return nil, ErrEnvNotFound
	}
	return persistEnvs(proj, updated)
}
