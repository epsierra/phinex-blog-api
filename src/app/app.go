package app

import (
	"github.com/epsierra/phinex-blog-api/src/auth"
	"github.com/epsierra/phinex-blog-api/src/blogs"
	"github.com/epsierra/phinex-blog-api/src/users"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AppSetup(db *gorm.DB) *fiber.App {
	app := fiber.New()

	blogService := blogs.NewBlogsService(db)
	blogController := blogs.NewBlogsController(blogService)
	blogController.RegisterRoutes(app)

	userService := users.NewUsersService(db)
	userController := users.NewUsersController(userService)
	userController.RegisterRoutes(app)

	authService := auth.NewAuthService(db)
	authController := auth.NewAuthController(authService)
	authController.RegisterRoutes(app)

	return app
}
