package respond

import "github.com/gofiber/fiber/v3"

type envelope struct {
	Ok   bool `json:"ok"`
	Data any  `json:"data"`
}

func OK(c fiber.Ctx, data any) error {
	return c.JSON(envelope{Ok: true, Data: data})
}

func Created(c fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(envelope{Ok: true, Data: data})
}

func Fail(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(envelope{Ok: false, Data: message})
}

func NoContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
