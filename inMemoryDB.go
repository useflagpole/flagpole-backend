package main

import "errors"

type FeatureFlagMapType map[string]bool
var FeatureFlagMap FeatureFlagMapType

func InitFeatureFlagMap() {
	if (FeatureFlagMap == nil) {
		FeatureFlagMap = make(FeatureFlagMapType)
	}
}

func (ffmap FeatureFlagMapType) AddFlag(name string, initvalue bool) error {
	if len(name) == 1 {
		return errors.New("Invalid feature flag name")
	}

	if _, exists := ffmap[name]; exists {
		return errors.New("Feature flag already exists")
	}
	ffmap[name] = initvalue
	return nil
}

func (ffmap FeatureFlagMapType) GetFlag(name string) (bool, error) {
	value, exists := ffmap[name]
	if (!exists) {
		return false, errors.New("Feature flag doesn't exist")
	}
	return value, nil
}

func (ffmap FeatureFlagMapType) SetFlag(name string, value bool) error {
	if _, exists := ffmap[name]; !exists {
		return errors.New("Feature flag doesn't exist")
	}
	ffmap[name] = value
	return nil
}
