package main

import (
	"log"

	"flagpole/src/config"
	"flagpole/src/database"
	"flagpole/src/routes"

	"github.com/gofiber/fiber/v3"
)

func main() {
	cfg := config.Get()

	if err := database.Init(cfg.DSN); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	app := fiber.New()
	routes.Setup(app)

	log.Fatal(app.Listen(":" + cfg.Port))
}
