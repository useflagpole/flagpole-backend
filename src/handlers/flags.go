package handlers

import (
	"encoding/json"

	"flagpole/src/controllers"
	"flagpole/src/models"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type flagPayload struct {
	FlagName string      `json:"flagName,omitempty"`
	Type     string      `json:"type,omitempty"`
	Value    interface{} `json:"value"`
}

type flagResponse struct {
	FlagName string          `json:"flagName"`
	Type     models.FlagType `json:"type"`
	Value    interface{}     `json:"value"`
}

type evaluationContext struct {
	Key        string                 `json:"key"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

type evaluateRequest struct {
	Context evaluationContext `json:"context"`
	Flags   []string          `json:"flags,omitempty"`
}

type evaluateResponse struct {
	Context evaluationContext           `json:"context"`
	Flags   map[string]models.FlagValue `json:"flags"`
}

// AddFlag godoc
// @Summary      Create a feature flag
// @Tags         flags
// @Accept       json
// @Produce      json
// @Param        body body flagPayload true "Flag definition"
// @Success      201
// @Failure      400 {object} response.ErrorResponse
// @Router       /flag [post]
func AddFlag(c fiber.Ctx) (int, response.APIResponse) {
	body := new(flagPayload)
	if err := json.Unmarshal(c.Body(), body); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}

	if err := controllers.AddFlag(body.FlagName, models.FlagType(body.Type), body.Value); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: err.Error()}
	}

	return fiber.StatusCreated, nil
}

// GetFlag godoc
// @Summary      Get a feature flag
// @Tags         flags
// @Produce      json
// @Param        flagname path string true "Flag name"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /flag/{flagname} [get]
func GetFlag(c fiber.Ctx) (int, response.APIResponse) {
	flagName := c.Params("flagname")
	if len(flagName) == 0 {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "missing flagname"}
	}

	fv, err := controllers.GetFlag(flagName)
	if err != nil {
		return fiber.StatusNotFound, response.ErrorResponse{Error: err.Error()}
	}

	return fiber.StatusOK, response.DataResponse{Data: flagResponse{FlagName: flagName, Type: fv.Type, Value: fv.Value}}
}

// SetFlag godoc
// @Summary      Update a feature flag's value
// @Tags         flags
// @Accept       json
// @Produce      json
// @Param        flagname path string true "Flag name"
// @Param        body body flagPayload true "New value"
// @Success      200
// @Failure      400 {object} response.ErrorResponse
// @Router       /flag/{flagname} [put]
func SetFlag(c fiber.Ctx) (int, response.APIResponse) {
	flagName := c.Params("flagname")
	if len(flagName) == 0 {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "missing flagname"}
	}

	body := new(flagPayload)
	if err := json.Unmarshal(c.Body(), body); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}

	if err := controllers.SetFlag(flagName, body.Value); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: err.Error()}
	}

	return fiber.StatusOK, nil
}

// EvaluateFlags godoc
// @Summary      Evaluate feature flags for a context
// @Tags         flags
// @Accept       json
// @Produce      json
// @Param        body body evaluateRequest true "Evaluation request"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Router       /flags/evaluate [post]
func EvaluateFlags(c fiber.Ctx) (int, response.APIResponse) {
	req := new(evaluateRequest)
	if err := json.Unmarshal(c.Body(), req); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}
	if len(req.Context.Key) == 0 {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "context.key is required"}
	}

	flags := controllers.EvaluateFlags(req.Flags)
	return fiber.StatusOK, response.DataResponse{Data: evaluateResponse{Context: req.Context, Flags: flags}}
}
