package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	httpSwagger "github.com/swaggo/http-swagger"
)

func HostSwaggerDocs(c fiber.Ctx) error {
	handler := adaptor.HTTPHandler(httpSwagger.Handler(
		httpSwagger.URL("/docs/doc.json"),
	))
	return handler(c)
}
