package routes

import "github.com/gofiber/fiber/v3"

func Setup(app *fiber.App) {
	app.Get("/status", func(c fiber.Ctx) error { return c.SendStatus(200) })

	app.Post("/signup", signup)
	app.Post("/login", login)

	app.Post("/flags/evaluate", evaluateFlags)
	app.Post("/flag", addFlag)
	app.Get("/flag/:flagname", getFlag)
	app.Put("/flag/:flagname", setFlag)
}
