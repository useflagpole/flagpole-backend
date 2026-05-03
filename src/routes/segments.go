package routes

import (
	"flagpole/src/handlers"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

func registerSegmentRoutes(api fiber.Router) {
	proj := "/organizations/:org_id/projects/:project_id"
	api.Get(proj+"/segments", response.Wrap(handlers.ListSegments))
	api.Post(proj+"/segments", response.Wrap(handlers.CreateSegment))
	api.Get(proj+"/segments/:segment_id", response.Wrap(handlers.GetSegment))
	api.Patch(proj+"/segments/:segment_id", response.Wrap(handlers.UpdateSegment))
	api.Delete(proj+"/segments/:segment_id", response.Wrap(handlers.DeleteSegment))
}
