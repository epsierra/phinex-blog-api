package app

import (
	"github.com/epsierra/phinex-blog-api/src/controllers"
	"github.com/epsierra/phinex-blog-api/src/middlewares"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AppSetup(db *gorm.DB) *fiber.App {
	app := fiber.New()

	// Setup Authorization middleware
	app.Use(middlewares.AuthMiddleware())

	blogController := controllers.NewBlogController(db)
	blogController.Register(app)
	return app
}
