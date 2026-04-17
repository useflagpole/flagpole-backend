package routes

import (
	"flagpole/src/config"
	"flagpole/src/handlers"
	"flagpole/src/middleware"
	"log"

	"github.com/gofiber/fiber/v3"
)

const apiVersion = "v1"

func Setup(app *fiber.App) {
	if config.Get().Env == "dev" {
		app.Get("/docs", func(c fiber.Ctx) error { return c.Redirect().To("/docs/index.html") })
		app.Get("/docs/*", handlers.HostSwaggerDocs)
		log.Println("Dev env detected. Serving docs")
	}

	api := app.Group("/api/" + apiVersion)

	api.Get("/status", func(c fiber.Ctx) error { return c.SendStatus(200) })

	registerAnonRoutes(api)

	guarded := api.Group("/", middleware.Auth)
	registerOrganizationRoutes(guarded)
	registerFlagRoutes(guarded)
}
