package blogs

import (
	"strconv"

	"github.com/epsierra/phinex-blog-api/src/middlewares"
	"github.com/epsierra/phinex-blog-api/src/models"
	"github.com/gofiber/fiber/v2"
)

// BlogsController handles HTTP requests for blogs
type BlogsController struct {
	service *BlogsService
}

// NewBlogsController creates a new BlogsController instance
func NewBlogsController(service *BlogsService) *BlogsController {
	return &BlogsController{
		service: service,
	}
}

// RegisterRoutes registers the blog-related routes to the Fiber app
func (c *BlogsController) RegisterRoutes(app *fiber.App) {
	// Guard
	app.Use("/blogs/*", middlewares.AuthenticatedGuard(c.service.db))
	app.Use("/users/*", middlewares.AuthenticatedGuard(c.service.db))
	app.Use("/following-blogs/*", middlewares.AuthenticatedGuard(c.service.db))
	app.Use("/pinned-blogs/*", middlewares.AuthenticatedGuard(c.service.db))
	app.Use("/comments/*", middlewares.AuthenticatedGuard(c.service.db))

	// Blog CRUD routes
	app.Post("/blogs", c.CreateBlog)                                 // Create a new blog post
	app.Get("/pinned-blogs", c.FindPinnedBlogs)                      // Get all pinned blogs
	app.Get("/blogs/:blogId", c.FindBlogById)                        // Get a blog post by its ID
	app.Get("/following-blogss", c.FindFollowingBlogs)               // Get all sessions blogs
	app.Get("/blogs", c.FindAllBlogs)                                // Get all blogs
	app.Put("/blogs/:blogId", c.UpdateBlog)                          // Update a blog post
	app.Put("/blogs/:blogId/likes", c.LikeBlog)                      // Like and/or unlike a blog post
	app.Delete("/blogs/:blogId", c.DeleteBlog)                       // Delete a blog post
	app.Get("/blogs/:blogId/follows/likes", c.FindLikesAndFollowers) // Get all likes and followed or followers that like a specific post

	// Comment-related routes
	app.Post("/blogs/:blogId/comments", c.AddComment)         // Add a comment to a blog post
	app.Get("/blogs/:blogId/comments", c.FindComments)        // Add a comment to a blog post
	app.Delete("/comments/:commentId", c.DeleteComment)       // Delete a comment
	app.Put("/comments/:commentId", c.UpdateComment)          // Update a comment
	app.Put("/comments/:commentId/likes", c.LikeComment)      // Like/unlike a comment
	app.Get("/comments/:commentId/likes", c.FindCommentLikes) // Get likes for a comment
	app.Get("/comments/:commentId/replies", c.FindReplies)    // Get replies to a comment

	// Reply-related routes
	app.Post("/comments/:commentId/replies", c.AddReply) // Add reply to a comment

	// User-related blog routes
	app.Get("/users/:userId/blogs", c.FindUserBlogs) // Get all user blogs

}

// @Summary Get all blogs
// @Description Get all blogs with pagination
// @Tags Blogs
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Success 200 {object} models.PaginatedResponse
// @Router /blogs [get]
// @Security ApiKeyAuth
func (c *BlogsController) FindAllBlogs(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))

	blogs, err := c.service.FindAll(currentUser, page, limit)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(blogs)
}

// @Summary Get pinned blogs
// @Description Get all pinned blogs with pagination
// @Tags Blogs
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Success 200 {object} models.PaginatedResponse
// @Router /blogs/pinned [get]
// @Security ApiKeyAuth
func (c *BlogsController) FindPinnedBlogs(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))

	blogs, err := c.service.FindPinnedBlogs(currentUser, page, limit)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(blogs)
}

// @Summary Get a single blog by ID
// @Description Get a single blog by its ID
// @Tags Blogs
// @Accept json
// @Produce json
// @Param blogId path string true "Blog ID"
// @Success 200 {object} BlogWithMeta
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /blogs/{blogId} [get]
// @Security ApiKeyAuth
func (c *BlogsController) FindBlogById(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	blogId := ctx.Params("blogId")

	blog, err := c.service.FindOne(blogId, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(blog)
}

// @Summary Get session blogs
// @Description Get blogs from followed users or the current user with pagination
// @Tags Blogs
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Success 200 {object} models.PaginatedResponse
// @Router /following-blogs [get]
// @Security ApiKeyAuth
func (c *BlogsController) FindFollowingBlogs(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))

	blogs, err := c.service.FindFollowingBlogs(currentUser, page, limit)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(blogs)
}

// @Summary Get user blogs
// @Description Get all blogs by a specific user with pagination
// @Tags Blogs
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Success 200 {object} models.PaginatedResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{userId}/blogs [get]
// @Security ApiKeyAuth
func (c *BlogsController) FindUserBlogs(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	userId := ctx.Params("userId")
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))

	blogs, err := c.service.FindUserBlogs(userId, currentUser, page, limit)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(blogs)
}

// @Summary Create a new blog
// @Description Create a new blog post
// @Tags Blogs
// @Accept json
// @Produce json
// @Param blog body CreateBlogDto true "Blog object to be created"
// @Success 201 {object} MutationResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /blogs [post]
// @Security ApiKeyAuth
func (c *BlogsController) CreateBlog(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	var dto CreateBlogDto
	if err := ctx.BodyParser(&dto); err != nil {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Invalid request body"}
	}

	response, err := c.service.Create(dto, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(response)
}

// @Summary Update an existing blog
// @Description Update a blog post by its ID
// @Tags Blogs
// @Accept json
// @Produce json
// @Param blogId path string true "Blog ID"
// @Param blog body UpdateBlogDto true "Blog object to be updated"
// @Success 202 {object} MutationResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /blogs/{blogId} [put]
// @Security ApiKeyAuth
func (c *BlogsController) UpdateBlog(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	blogId := ctx.Params("blogId")
	var dto UpdateBlogDto
	if err := ctx.BodyParser(&dto); err != nil {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Invalid request body"}
	}

	response, err := c.service.Update(blogId, dto, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusAccepted).JSON(response)
}

// @Summary Delete a blog
// @Description Delete a blog post by its ID
// @Tags Blogs
// @Accept json
// @Produce json
// @Param blogId path string true "Blog ID"
// @Success 204 {object} map[string]string
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /blogs/{blogId} [delete]
// @Security ApiKeyAuth
func (c *BlogsController) DeleteBlog(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	blogId := ctx.Params("blogId")

	response, err := c.service.Delete(blogId, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusNonAuthoritativeInformation).JSON(response)
}

// @Summary Like or unlike a blog
// @Description Like and/or unlike a blog post by its ID
// @Tags Blogs
// @Accept json
// @Produce json
// @Param blogId path string true "Blog ID"
// @Success 200 {object} LikeResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /blogs/{blogId}/likes [put]
// @Security ApiKeyAuth
func (c *BlogsController) LikeBlog(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	blogId := ctx.Params("blogId")

	response, err := c.service.LikeBlog(blogId, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusAccepted).JSON(response)
}

// @Summary Add a comment to a blog
// @Description Add a comment to a blog post by its ID
// @Tags Comments
// @Accept json
// @Produce json
// @Param blogId path string true "Blog ID"
// @Param comment body CreateCommentDto true "Comment object to be created"
// @Success 201 {object} MutationResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /blogs/{blogId}/comments [post]
// @Security ApiKeyAuth
func (c *BlogsController) AddComment(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	blogId := ctx.Params("blogId")
	var dto CreateCommentDto
	if err := ctx.BodyParser(&dto); err != nil {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Invalid request body"}
	}

	response, err := c.service.AddComment(blogId, dto, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(response)
}

// @Summary Delete a comment
// @Description Delete a comment by its ID
// @Tags Comments
// @Accept json
// @Produce json
// @Param commentId path string true "Comment ID"
// @Success 204 {object} map[string]string
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /comments/{commentId} [delete]
// @Security ApiKeyAuth
func (c *BlogsController) DeleteComment(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	commentId := ctx.Params("commentId")

	response, err := c.service.DeleteComment(commentId, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusNonAuthoritativeInformation).JSON(response)
}

// @Summary Update a comment
// @Description Update a comment by its ID
// @Tags Comments
// @Accept json
// @Produce json
// @Param commentId path string true "Comment ID"
// @Param comment body CreateCommentDto true "Comment object to be updated"
// @Success 202 {object} MutationResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /comments/{commentId} [put]
// @Security ApiKeyAuth
func (c *BlogsController) UpdateComment(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	commentId := ctx.Params("commentId")
	var dto CreateCommentDto
	if err := ctx.BodyParser(&dto); err != nil {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Invalid request body"}
	}

	response, err := c.service.UpdateComment(commentId, dto, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusAccepted).JSON(response)
}

// @Summary Like or unlike a comment
// @Description Like and/or unlike a comment by its ID
// @Tags Comments
// @Accept json
// @Produce json
// @Param commentId path string true "Comment ID"
// @Success 200 {object} LikeResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /comments/{commentId}/likes [put]
// @Security ApiKeyAuth
func (c *BlogsController) LikeComment(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	commentId := ctx.Params("commentId")

	response, err := c.service.LikeComment(commentId, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusAccepted).JSON(response)
}

// @Summary Get comment likes
// @Description Get all likes for a specific comment
// @Tags Comments
// @Accept json
// @Produce json
// @Param commentId path string true "Comment ID"
// @Success 200 {array} models.Like
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /comments/{commentId}/likes [get]
// @Security ApiKeyAuth
func (c *BlogsController) FindCommentLikes(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	commentId := ctx.Params("commentId")

	likes, err := c.service.FindCommentLikes(commentId, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(likes)
}

// @Summary Get comments for a blog
// @Description Get all comments for a specific blog with pagination
// @Tags Comments
// @Accept json
// @Produce json
// @Param blogId path string true "Blog ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Success 200 {object} models.PaginatedResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /blogs/{blogId}/comments [get]
// @Security ApiKeyAuth
func (c *BlogsController) FindComments(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	blogId := ctx.Params("blogId")
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))

	comments, err := c.service.FindComments(blogId, currentUser, page, limit)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(comments)
}

// @Summary Get replies for a comment
// @Description Get all replies for a specific comment with pagination
// @Tags Comments
// @Accept json
// @Produce json
// @Param commentId path string true "Comment ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Success 200 {object} models.PaginatedResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /comments/{commentId}/replies [get]
// @Security ApiKeyAuth
func (c *BlogsController) FindReplies(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	commentId := ctx.Params("commentId")
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))

	replies, err := c.service.FindReplies(commentId, currentUser, page, limit)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(replies)
}

// @Summary Add a reply to a comment
// @Description Add a reply to a comment by its ID
// @Tags Comments
// @Accept json
// @Produce json
// @Param commentId path string true "Comment ID"
// @Param reply body CreateReplyDto true "Reply object to be created"
// @Success 201 {object} MutationResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /comments/{commentId}/replies [post]
// @Security ApiKeyAuth
func (c *BlogsController) AddReply(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	commentId := ctx.Params("commentId")
	var dto CreateReplyDto
	if err := ctx.BodyParser(&dto); err != nil {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "Invalid request body"}
	}

	response, err := c.service.AddReply(commentId, dto, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(response)
}

// @Summary Get likes and followers for a blog
// @Description Get all likes and followed or followers that like a specific blog
// @Tags Blogs
// @Accept json
// @Produce json
// @Param blogId path string true "Blog ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /blogs/{blogId}/follows/likes [get]
// @Security ApiKeyAuth
func (c *BlogsController) FindLikesAndFollowers(ctx *fiber.Ctx) error {
	currentUser := ctx.Locals("user").(models.ICurrentUser)
	blogId := ctx.Params("blogId")

	response, err := c.service.FindLikesAndFollowers(blogId, currentUser)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(response)
}
