package routes

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

func addFlag(c fiber.Ctx) error {
	body := new(flagPayload)
	if err := json.Unmarshal(c.Body(), body); err != nil {
		return c.Status(400).SendString("couldn't parse body")
	}

	if err := controllers.AddFlag(body.FlagName, models.FlagType(body.Type), body.Value); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	return c.SendStatus(200)
}

func getFlag(c fiber.Ctx) error {
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

func setFlag(c fiber.Ctx) error {
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

func evaluateFlags(c fiber.Ctx) error {
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
