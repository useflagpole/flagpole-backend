package controllers

import (
	"errors"
	"fmt"
	"time"

	"flagpole/src/dal"
	"flagpole/src/models"
	"flagpole/src/pkg/crypto"
)

const MAX_ACTIVE_SDK_KEYS_PER_SLOT = 5

var (
	ErrSDKKeyNotFound       = errors.New("sdk key not found")
	ErrSDKKeyNameRequired   = errors.New("key name is required")
	ErrSDKKeyNameTooLong    = errors.New("key name must be at most " + fmt.Sprint(models.SDKKeyNameMaxLen) + " characters")
	ErrSDKKeyTypeInvalid    = errors.New("key type must be 'server' or 'client'")
	ErrSDKKeyLimitReached   = errors.New("maximum active keys for this environment and type reached")
	ErrSDKKeyAlreadyRevoked = errors.New("sdk key is already revoked")
	ErrSDKKeyEnvNotFound    = errors.New("environment not found in this project")
)

type SDKKeyCreatedDTO struct {
	ID              uint      `json:"id"`
	ProjectID       uint      `json:"projectId"`
	EnvironmentID   uint      `json:"environmentId"`
	EnvironmentName string    `json:"environmentName"`
	KeyType         string    `json:"type"`
	Name            string    `json:"name"`
	Key             string    `json:"key"`
	CreatedAt       time.Time `json:"createdAt"`
}

type SDKKeyDTO struct {
	ID              uint       `json:"id"`
	ProjectID       uint       `json:"projectId"`
	EnvironmentID   uint       `json:"environmentId"`
	EnvironmentName string     `json:"environmentName"`
	KeyType         string     `json:"type"`
	Name            string     `json:"name"`
	KeyHint         string     `json:"keyHint"` // prefix + last 4 chars, e.g. "fp_srv_…a1b2"
	RevokedAt       *time.Time `json:"revokedAt"`
	LastUsedAt      *time.Time `json:"lastUsedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
}

func CreateSDKKey(projectID, environmentID uint, keyType, name string) (*SDKKeyCreatedDTO, error) {
	if name == "" {
		return nil, ErrSDKKeyNameRequired
	}
	if len(name) > models.SDKKeyNameMaxLen {
		return nil, ErrSDKKeyNameTooLong
	}
	if keyType != models.SDKKeyTypeServer && keyType != models.SDKKeyTypeClient {
		return nil, ErrSDKKeyTypeInvalid
	}

	env, err := dal.Environment.GetByID(environmentID)
	if err != nil || env.ProjectID != projectID {
		return nil, ErrSDKKeyEnvNotFound
	}

	count, err := dal.SDKKey.CountActive(projectID, environmentID, keyType)
	if err != nil {
		return nil, err
	}
	if count >= MAX_ACTIVE_SDK_KEYS_PER_SLOT {
		return nil, ErrSDKKeyLimitReached
	}

	prefix := models.SDKKeyPrefixServer
	if keyType == models.SDKKeyTypeClient {
		prefix = models.SDKKeyPrefixClient
	}
	rawKey, err := crypto.GenerateSDKKey(prefix)
	if err != nil {
		return nil, err
	}

	k := &models.SDKKey{
		ProjectID:     projectID,
		EnvironmentID: environmentID,
		KeyType:       keyType,
		Name:          name,
		Key:           rawKey,
	}
	if err := dal.SDKKey.Create(k); err != nil {
		return nil, err
	}

	return &SDKKeyCreatedDTO{
		ID:              k.ID,
		ProjectID:       k.ProjectID,
		EnvironmentID:   k.EnvironmentID,
		EnvironmentName: env.Name,
		KeyType:         k.KeyType,
		Name:            k.Name,
		Key:             rawKey,
		CreatedAt:       k.CreatedAt,
	}, nil
}

func ListSDKKeys(projectID, environmentID uint) ([]SDKKeyDTO, error) {
	keys, err := dal.SDKKey.ListByProject(projectID, environmentID)
	if err != nil {
		return nil, err
	}
	dtos := make([]SDKKeyDTO, len(keys))
	for i, k := range keys {
		hint := ""
		if len(k.Key) > 4 {
			prefix := models.SDKKeyPrefixServer
			if k.KeyType == models.SDKKeyTypeClient {
				prefix = models.SDKKeyPrefixClient
			}
			hint = prefix + k.Key[len(k.Key)-4:]
		}
		dtos[i] = SDKKeyDTO{
			ID:              k.ID,
			ProjectID:       k.ProjectID,
			EnvironmentID:   k.EnvironmentID,
			EnvironmentName: k.Environment.Name,
			KeyType:         k.KeyType,
			Name:            k.Name,
			KeyHint:         hint,
			RevokedAt:       k.RevokedAt,
			LastUsedAt:      k.LastUsedAt,
			CreatedAt:       k.CreatedAt,
		}
	}
	return dtos, nil
}

func RevokeSDKKey(keyID, projectID uint) (keyName string, err error) {
	k, dbErr := dal.SDKKey.GetByID(keyID, projectID)
	if dbErr != nil {
		return "", ErrSDKKeyNotFound
	}
	if k.RevokedAt != nil {
		return "", ErrSDKKeyAlreadyRevoked
	}
	return k.Name, dal.SDKKey.Revoke(k)
}

func RevealSDKKey(keyID, projectID uint) (rawKey, keyName string, err error) {
	k, dbErr := dal.SDKKey.GetByID(keyID, projectID)
	if dbErr != nil {
		return "", "", ErrSDKKeyNotFound
	}
	if k.RevokedAt != nil {
		return "", "", ErrSDKKeyAlreadyRevoked
	}
	return k.Key, k.Name, nil
}
