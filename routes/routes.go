package routes

import (
	"github.com/gofiber/fiber/v2"
	"go-login-auth/controllers"
	"go-login-auth/middleware"
)

func Setup(app *fiber.App) {
	app.Post("/user/login", controllers.Login)

	/* Accept Only with Access Token */
	app.Use(middleware.UserAuthorization)
	app.Get("/user/info", controllers.UserInfo)
	app.Get("/user/refresh", controllers.Refresh)
	app.Get("/user/logout", controllers.Logout)
}
