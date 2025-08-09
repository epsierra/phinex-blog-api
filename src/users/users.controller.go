package users

import (
	"strconv"

	"github.com/epsierra/phinex-blog-api/src/middlewares"
	"github.com/epsierra/phinex-blog-api/src/models"
	"github.com/gofiber/fiber/v2"
)

// UsersController handles HTTP requests for users
type UsersController struct {
	service *UsersService
}

// NewUsersController creates a new UsersController instance
func NewUsersController(service *UsersService) *UsersController {
	return &UsersController{
		service: service,
	}
}

// RegisterRoutes registers the user-related routes to the Fiber app
func (c *UsersController) RegisterRoutes(app *fiber.App) {
	// Guard
	app.Use("/users/*", middlewares.AuthenticatedGuard(c.service.db))

	app.Post("/users", c.CreateUser)
	app.Get("/users", c.FindAllUsers)
	app.Get("/users/:userId", c.FindUserById)
	app.Put("/users/:userId", c.UpdateUser)
	app.Delete("/users/:userId", c.DeleteUser)
	app.Post("/users/follows", c.FollowUnfollowUser)
	app.Get("/users/:userId/unfollowings", c.FindUsersNotFollowing)
	app.Get("/users/:userId/followers", c.FindUserFollowers)
	app.Get("/users/:userId/followings", c.FindUserFollowings)
}

// @Summary Create a new user
// @Description Create a new user with the provided details
// @Tags Users
// @Accept json
// @Produce json
// @Param user body CreateUserDto true "User object to be created"
// @Success 201 {object} UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users [post]
func (c *UsersController) CreateUser(ctx *fiber.Ctx) error {
	var dto CreateUserDto
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	if err := ctx.BodyParser(&dto); err != nil {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Invalid request body"}
	}

	response, err := c.service.CreateUser(dto, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(response)
}

// @Summary Get all users
// @Description Get all users with pagination and optional search
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Param search query string false "Search term to filter users by full name, email, username, or bio"
// @Success 200 {object} models.PaginatedResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users [get]
// @Security ApiKeyAuth
func (c *UsersController) FindAllUsers(ctx *fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	search := ctx.Query("search", "")
	currentUser := ctx.Locals("user").(models.ICurrentUser)

	users, err := c.service.FindAllUsers(page, limit, search, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(users)
}

// @Summary Get user by ID
// @Description Get a single user by their ID
// @Tags Users
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} models.User
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{userId} [get]
// @Security ApiKeyAuth
func (c *UsersController) FindUserById(ctx *fiber.Ctx) error {
	userId := ctx.Params("userId")
	currentUser := ctx.Locals("user").(models.ICurrentUser)

	user, err := c.service.FindUserById(userId, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(user)
}

// @Summary Update a user
// @Description Update a user's details by their ID
// @Tags Users
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param user body UpdateUserDto true "User object with updated fields"
// @Success 202 {object} UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{userId} [put]
// @Security ApiKeyAuth
func (c *UsersController) UpdateUser(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	userId := ctx.Params("userId")
	var dto UpdateUserDto
	if err := ctx.BodyParser(&dto); err != nil {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Invalid request body"}
	}

	response, err := c.service.UpdateUser(userId, dto, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusAccepted).JSON(response)
}

// @Summary Delete a user
// @Description Delete a user by their ID
// @Tags Users
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 204 {object} UserResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{userId} [delete]
// @Security ApiKeyAuth
func (c *UsersController) DeleteUser(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	userId := ctx.Params("userId")
	response, err := c.service.DeleteUser(userId, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusNonAuthoritativeInformation).JSON(response)
}

// @Summary Follow or unfollow a user
// @Description Allows a user to follow or unfollow another user
// @Tags Users
// @Accept json
// @Produce json
// @Param follow body FollowUnfollowDto true "Follow/Unfollow object"
// @Success 200 {object} FollowResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 422 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/follows [post]
// @Security ApiKeyAuth
func (c *UsersController) FollowUnfollowUser(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	var dto FollowUnfollowDto
	if err := ctx.BodyParser(&dto); err != nil {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Invalid request body"}
	}

	response, err := c.service.FollowUnfollowUser(dto, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(response)
}

// @Summary Get users not following
// @Description Retrieves users who are not followed by a specific user with pagination and optional search
// @Tags Users
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Param search query string false "Search term to filter users by full name, email, username, or bio"
// @Success 200 {object} models.PaginatedResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{userId}/unfollowings [get]
// @Security ApiKeyAuth
func (c *UsersController) FindUsersNotFollowing(ctx *fiber.Ctx) error {
	userId := ctx.Params("userId")
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	search := ctx.Query("search", "")
	currentUser := ctx.Locals("user").(models.ICurrentUser)

	users, err := c.service.FindUsersNotFollowing(userId, page, limit, search, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(users)
}

// @Summary Get user followers
// @Description Retrieves the followers of a specific user with pagination and optional search
// @Tags Users
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Param search query string false "Search term to filter users by full name, email, username, or bio"
// @Success 200 {object} models.PaginatedResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{userId}/followers [get]
// @Security ApiKeyAuth
func (c *UsersController) FindUserFollowers(ctx *fiber.Ctx) error {
	userId := ctx.Params("userId")
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	search := ctx.Query("search", "")
	currentUser := ctx.Locals("user").(models.ICurrentUser)

	users, err := c.service.FindUserFollowers(userId, page, limit, search, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(users)
}

// @Summary Get user followings
// @Description Retrieves a user's followings with pagination and optional search
// @Tags Users
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Param search query string false "Search term to filter users by full name, email, username, or bio"
// @Success 200 {object} models.PaginatedResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{userId}/followings [get]
// @Security ApiKeyAuth
func (c *UsersController) FindUserFollowings(ctx *fiber.Ctx) error {
	userId := ctx.Params("userId")
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	search := ctx.Query("search", "")
	currentUser := ctx.Locals("user").(models.ICurrentUser)

	users, err := c.service.FindUserFollowings(userId, page, limit, search, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(users)
}
