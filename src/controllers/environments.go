package controllers

import (
	"errors"
	"fmt"

	"flagpole/src/dal"
	"flagpole/src/models"
)

const MAX_ENVIRONMENTS = 5

var (
	ErrMaxEnvironments  = errors.New("maximum of " + fmt.Sprint(MAX_ENVIRONMENTS) + " environments reached")
	ErrEnvAlreadyExists = errors.New("environment already exists")
	ErrEnvNotFound      = errors.New("environment not found")
	ErrEnvProtected     = errors.New("production environment is protected")
)

func ListEnvironments(projectID uint) ([]string, error) {
	envs, err := dal.Environment.ListByProject(projectID)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(envs))
	for i, e := range envs {
		names[i] = e.Name
	}
	return names, nil
}

func CreateEnvironment(projectID uint, name string) ([]string, error) {
	if err := models.ValidateEnvironmentName(name); err != nil {
		return nil, err
	}
	if dal.Environment.NameExists(projectID, name) {
		return nil, ErrEnvAlreadyExists
	}
	envs, err := ListEnvironments(projectID)
	if err != nil {
		return nil, err
	}
	if len(envs) >= MAX_ENVIRONMENTS {
		return nil, ErrMaxEnvironments
	}
	env := &models.Environment{
		ProjectID: projectID,
		Name:      name,
	}
	if err := dal.Environment.Create(env); err != nil {
		return nil, err
	}
	envs = append(envs, name)
	return envs, nil
}

func RenameEnvironment(projectID uint, oldName, newName string) ([]string, error) {
	if oldName == "production" {
		return nil, ErrEnvProtected
	}
	if err := models.ValidateEnvironmentName(newName); err != nil {
		return nil, err
	}
	env, err := dal.Environment.GetByName(projectID, oldName)
	if err != nil {
		return nil, ErrEnvNotFound
	}
	if dal.Environment.NameExists(projectID, newName) {
		return nil, ErrEnvAlreadyExists
	}
	env.Name = newName
	if err := dal.Environment.Save(env); err != nil {
		return nil, err
	}
	return ListEnvironments(projectID)
}

func DeleteEnvironment(projectID uint, name string) ([]string, error) {
	if name == "production" {
		return nil, ErrEnvProtected
	}
	env, err := dal.Environment.GetByName(projectID, name)
	if err != nil {
		return nil, ErrEnvNotFound
	}

	if err := dal.FlagEnvConfig.DeleteByEnv(projectID, name); err != nil {
		return nil, err
	}

	flags, err := dal.FeatureFlag.ListByProject(projectID)
	if err != nil {
		return nil, err
	}
	for _, flag := range flags {
		if err := dal.FlagEnvOverride.RemoveByEnv(flag.ID, name); err != nil {
			return nil, err
		}
	}

	if err := dal.Environment.Delete(env); err != nil {
		return nil, err
	}

	return ListEnvironments(projectID)
}
