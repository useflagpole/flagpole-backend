package routes

import (
	"flagpole/src/handlers"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

func registerSDKKeyRoutes(api fiber.Router) {
	proj := "/organizations/:org_id/projects/:project_id"
	api.Get(proj+"/sdk-keys", response.Wrap(handlers.ListSDKKeys))
	api.Post(proj+"/sdk-keys", response.Wrap(handlers.CreateSDKKey))
	api.Delete(proj+"/sdk-keys/:key_id", response.Wrap(handlers.RevokeSDKKey))
	api.Get(proj+"/sdk-keys/:key_id/reveal", response.Wrap(handlers.RevealSDKKey))
}
