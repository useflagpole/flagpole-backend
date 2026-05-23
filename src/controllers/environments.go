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

func ListEnvironments(projectID uint) ([]models.Environment, error) {
	return dal.Environment.ListByProject(projectID)
}

func CreateEnvironment(projectID uint, name string) ([]models.Environment, error) {
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
	return ListEnvironments(projectID)
}

func RenameEnvironment(projectID uint, envID uint, newName string) ([]models.Environment, error) {
	env, err := dal.Environment.GetByID(envID)
	if err != nil || env.ProjectID != projectID {
		return nil, ErrEnvNotFound
	}
	if env.Name == "production" {
		return nil, ErrEnvProtected
	}
	if err := models.ValidateEnvironmentName(newName); err != nil {
		return nil, err
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

func DeleteEnvironment(projectID uint, envID uint) ([]models.Environment, error) {
	env, err := dal.Environment.GetByID(envID)
	if err != nil || env.ProjectID != projectID {
		return nil, ErrEnvNotFound
	}
	if env.Name == "production" {
		return nil, ErrEnvProtected
	}

	if err := dal.FlagEnvConfig.DeleteByEnvID(projectID, env.ID); err != nil {
		return nil, err
	}

	flags, err := dal.FeatureFlag.ListByProject(projectID)
	if err != nil {
		return nil, err
	}
	for _, flag := range flags {
		if err := dal.FlagEnvOverride.RemoveByEnvID(flag.ID, env.ID); err != nil {
			return nil, err
		}
	}

	if err := dal.Environment.Delete(env); err != nil {
		return nil, err
	}

	return ListEnvironments(projectID)
}
