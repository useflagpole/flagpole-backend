package response

import "github.com/gofiber/fiber/v3"

type HandlerFunc func(fiber.Ctx) (int, APIResponse)

func Wrap(h HandlerFunc) fiber.Handler {
	return func(c fiber.Ctx) error {
		status, r := h(c)
		if r == nil {
			return c.SendStatus(status)
		}
		return c.Status(status).JSON(r)
	}
}
