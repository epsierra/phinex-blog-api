package auth

import (
	"github.com/gofiber/fiber/v2"
)

// AuthController handles HTTP requests for authentication
type AuthController struct {
	service *AuthService
}

// NewAuthController creates a new AuthController instance
func NewAuthController(service *AuthService) *AuthController {
	return &AuthController{
		service: service,
	}
}

// RegisterRoutes registers the authentication-related routes to the Fiber app
func (c *AuthController) RegisterRoutes(app *fiber.App) {
	app.Post("/auth/token", c.GetTokenByEmail)
}

// @Summary Get JWT token by email
// @Description Generates a JWT token for a verified user by email.
// @Tags Auth
// @Accept json
// @Produce json
// @Param email body GetTokenDto true "User email for token generation"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/token [post]
func (c *AuthController) GetTokenByEmail(ctx *fiber.Ctx) error {
	var dto GetTokenDto
	if err := ctx.BodyParser(&dto); err != nil {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Invalid request body"}
	}

	response, err := c.service.GetTokenByEmail(dto.Email)
	if err != nil {
		return err
	}
	return ctx.JSON(response)
}
