package routes

import (
	"flagpole/src/config"
	"flagpole/src/handlers"
	"flagpole/src/middleware"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

const apiVersion = "v1"

func Setup(app *fiber.App) {
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{config.Get().AllowOrigin},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
	}))

	if config.Get().Env == "dev" {
		app.Get("/docs", func(c fiber.Ctx) error { return c.Redirect().To("/docs/index.html") })
		app.Get("/docs/*", handlers.HostSwaggerDocs)
		log.Println("Dev env detected. Serving docs")
	}

	api := app.Group("/api/" + apiVersion)

	api.Get("/status", func(c fiber.Ctx) error { return c.SendStatus(200) })

	registerAnonRoutes(api)

	guarded := api.Group("/", middleware.Auth)
	registerUserRoutes(guarded)
	registerOrganizationRoutes(guarded)
	registerProjectRoutes(guarded)
	registerFlagRoutes(guarded)
}
