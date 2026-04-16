package main

import (
	"github.com/gofiber/fiber/v3"
)

func AddFlagRoute(c fiber.Ctx) error {
	flagBody := new(FlagPayload)
	if err := c.Bind().Body(flagBody); err != nil {
		c.SendString("Couldn't parse body")
		return c.SendStatus(500)
	}

	err := FeatureFlagMap.AddFlag(flagBody.FlagName, flagBody.Value)
	if err != nil {
		c.SendString(err.Error())
		return c.SendStatus(400)
	}

	return c.SendStatus(200)
}

func GetFlagRoute(c fiber.Ctx) error {
	flagName := c.Params("flagname")
	if len(flagName) == 0 {
		c.SendString("Missing flagname")
		return c.SendStatus(400)
	}

	flagValue, err := FeatureFlagMap.GetFlag(flagName)
	if err != nil {
		return c.SendStatus(400)
	}

	payload, err := GetJSONEncodedFlag(flagName, flagValue)
	if err != nil {
		return c.SendStatus(500)
	}

	return c.SendString(payload)
}

func SetFlagRoute(c fiber.Ctx) error {
	flagName := c.Params("flagname")
	if len(flagName) == 0 {
		c.SendString("Missing flagname")
		return c.SendStatus(400)
	}

	flagBody := new(FlagPayload)
	if err := c.Bind().Body(flagBody); err != nil {
		c.SendString("Couldn't parse body")
		return c.SendStatus(500)
	}

	err := FeatureFlagMap.SetFlag(flagName, flagBody.Value)
	if err != nil {
		c.SendString(err.Error())
		return c.SendStatus(400)
	}

	return c.SendStatus(200)
}
