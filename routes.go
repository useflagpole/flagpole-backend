package main

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

func AddFlagRoute(c fiber.Ctx) error {
	flagBody := new(FlagPayload)
	if err := json.Unmarshal(c.Body(), flagBody); err != nil {
		return c.Status(400).SendString("couldn't parse body")
	}

	err := FeatureFlagMap.AddFlag(flagBody.FlagName, FlagType(flagBody.Type), flagBody.Value)
	if err != nil {
		return c.Status(400).SendString(err.Error())
	}

	return c.SendStatus(200)
}

func GetFlagRoute(c fiber.Ctx) error {
	flagName := c.Params("flagname")
	if len(flagName) == 0 {
		return c.Status(400).SendString("missing flagname")
	}

	fv, err := FeatureFlagMap.GetFlag(flagName)
	if err != nil {
		return c.Status(404).SendString(err.Error())
	}

	payload, err := GetJSONEncodedFlag(flagName, fv)
	if err != nil {
		return c.SendStatus(500)
	}

	return c.SendString(payload)
}

func SetFlagRoute(c fiber.Ctx) error {
	flagName := c.Params("flagname")
	if len(flagName) == 0 {
		return c.Status(400).SendString("missing flagname")
	}

	flagBody := new(FlagPayload)
	if err := json.Unmarshal(c.Body(), flagBody); err != nil {
		return c.Status(400).SendString("couldn't parse body")
	}

	if err := FeatureFlagMap.SetFlag(flagName, flagBody.Value); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	return c.SendStatus(200)
}
