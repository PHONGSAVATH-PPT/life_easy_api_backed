package lifeweb

import (
	"life_easy_api/api/controllers/life"
	"life_easy_api/api/middleware"

	"github.com/gofiber/fiber/v2"
)

func Register(
	app *fiber.App,
	controller *life.Controller,
	secret string,
) {
	api := app.Group("/api/v1")

	auth := api.Group("/auth")
	auth.Post("/register", controller.Register)
	auth.Post("/login", controller.Login)

	protected := api.Group("", middleware.Auth(secret))

	protected.Get("/workspace", controller.Workspace)

	protected.Put("/entries/:id", controller.Save)
	protected.Delete("/entries/:id", controller.Remove)
	protected.Post("/entries/:id/toggle", controller.Toggle)
	protected.Post("/bills/:id/pay", controller.Pay)

	protected.Post("/families", controller.CreateFamily)
	protected.Post("/families/:id/members", controller.AddMember)
	protected.Delete(
		"/families/:id/members/:userID",
		controller.RemoveMember,
	)
}
