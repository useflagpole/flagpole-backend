// @title           flagpole API
// @version         1.0
// @description     Feature flag management API
// @host            localhost:4000
// @BasePath        /api/v1

package main

import (
	"log"

	"flagpole/src/config"
	"flagpole/src/database"
	_ "flagpole/src/docs"
	"flagpole/src/routes"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func main() {
	cfg := config.Get()

	if err := database.Init(cfg.DSN); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	app := fiber.New()
	if cfg.Env == "dev" {
		app.Use(logger.New())
	}
	routes.Setup(app)

	log.Fatal(app.Listen(":" + cfg.Port))
}
