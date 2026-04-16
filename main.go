package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
)

var flagReader Flags

func main() {
	flagReader.parseFlags()
	InitFeatureFlagMap()

	app := fiber.New()

	app.Get("/status", func(c fiber.Ctx) error {
		return c.SendStatus(200)
	})

	app.Post("/signup", SignupRoute)
	app.Post("/login", LoginRoute)

	app.Post("/flags/evaluate", EvaluateRoute)
	app.Post("/flag", AddFlagRoute)
	app.Get("/flag/:flagname", GetFlagRoute)
	app.Put("/flag/:flagname", SetFlagRoute)

	log.Fatal(app.Listen(":" + flagReader.port))
}
