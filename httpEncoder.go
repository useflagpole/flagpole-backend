package main

import (
	"encoding/json"
	"errors"
)

type FlagPayload struct {
	FlagName string      `json:"flagName,omitempty"`
	Type     string      `json:"type,omitempty"`
	Value    interface{} `json:"value"`
}

type FlagResponse struct {
	FlagName string      `json:"flagName"`
	Type     FlagType    `json:"type"`
	Value    interface{} `json:"value"`
}

type EvaluationContext struct {
	Key        string                 `json:"key"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

type EvaluateRequest struct {
	Context EvaluationContext `json:"context"`
	Flags   []string          `json:"flags,omitempty"`
}

type EvaluateResponse struct {
	Context EvaluationContext  `json:"context"`
	Flags   map[string]FlagValue `json:"flags"`
}

func GetJSONEncodedFlag(name string, fv FlagValue) (string, error) {
	if len(name) < 1 {
		return "", errors.New("invalid flag name")
	}
	payload, err := json.Marshal(FlagResponse{FlagName: name, Type: fv.Type, Value: fv.Value})
	if err != nil {
		return "", err
	}
	return string(payload), nil
}
