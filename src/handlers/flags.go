package handlers

import (
	"encoding/json"

	"flagpole/src/controllers"
	"flagpole/src/models"

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
// @Produce      plain
// @Param        body body flagPayload true "Flag definition"
// @Success      200
// @Failure      400 {string} string "bad request"
// @Router       /flag [post]
func AddFlag(c fiber.Ctx) error {
	body := new(flagPayload)
	if err := json.Unmarshal(c.Body(), body); err != nil {
		return c.Status(400).SendString("couldn't parse body")
	}

	if err := controllers.AddFlag(body.FlagName, models.FlagType(body.Type), body.Value); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	return c.SendStatus(200)
}

// GetFlag godoc
// @Summary      Get a feature flag
// @Tags         flags
// @Produce      json
// @Param        flagname path string true "Flag name"
// @Success      200 {object} flagResponse
// @Failure      400 {string} string "missing flagname"
// @Failure      404 {string} string "flag not found"
// @Router       /flag/{flagname} [get]
func GetFlag(c fiber.Ctx) error {
	flagName := c.Params("flagname")
	if len(flagName) == 0 {
		return c.Status(400).SendString("missing flagname")
	}

	fv, err := controllers.GetFlag(flagName)
	if err != nil {
		return c.Status(404).SendString(err.Error())
	}

	payload, err := json.Marshal(flagResponse{FlagName: flagName, Type: fv.Type, Value: fv.Value})
	if err != nil {
		return c.SendStatus(500)
	}

	return c.Send(payload)
}

// SetFlag godoc
// @Summary      Update a feature flag's value
// @Tags         flags
// @Accept       json
// @Produce      plain
// @Param        flagname path string true "Flag name"
// @Param        body body flagPayload true "New value"
// @Success      200
// @Failure      400 {string} string "bad request"
// @Router       /flag/{flagname} [put]
func SetFlag(c fiber.Ctx) error {
	flagName := c.Params("flagname")
	if len(flagName) == 0 {
		return c.Status(400).SendString("missing flagname")
	}

	body := new(flagPayload)
	if err := json.Unmarshal(c.Body(), body); err != nil {
		return c.Status(400).SendString("couldn't parse body")
	}

	if err := controllers.SetFlag(flagName, body.Value); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	return c.SendStatus(200)
}

// EvaluateFlags godoc
// @Summary      Evaluate feature flags for a context
// @Tags         flags
// @Accept       json
// @Produce      json
// @Param        body body evaluateRequest true "Evaluation request"
// @Success      200 {object} evaluateResponse
// @Failure      400 {string} string "bad request"
// @Router       /flags/evaluate [post]
func EvaluateFlags(c fiber.Ctx) error {
	req := new(evaluateRequest)
	if err := json.Unmarshal(c.Body(), req); err != nil {
		return c.Status(400).SendString("couldn't parse body")
	}
	if len(req.Context.Key) == 0 {
		return c.Status(400).SendString("context.key is required")
	}

	flags := controllers.EvaluateFlags(req.Flags)
	return c.JSON(evaluateResponse{Context: req.Context, Flags: flags})
}
