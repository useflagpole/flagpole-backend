package main

import (
	"encoding/json"
	"errors"
)

type FlagPayload struct {
	FlagName string `json:"flagName"`
	Value bool `json:"value"`
}

func GetJSONEncodedFlag(flagName string, value bool) (string, error) {
	if (len(flagName) < 1) {
		return "", errors.New("Invalid flag name")
	}

	flagPayload := FlagPayload{
		FlagName: flagName,
		Value: value,
	}

	jsonPayload, err := json.Marshal(flagPayload);
	if err != nil {
	  panic(err)
	}

  return string(jsonPayload), nil
}
