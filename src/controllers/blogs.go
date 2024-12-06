package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/epsierra/phinex-blog-api/src/models" // import your models package
	"github.com/epsierra/phinex-blog-api/src/utils"
	"github.com/google/uuid"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type blogController struct {
	db *gorm.DB
}

// NewBlogController creates a new blog controller with a database connection
func NewBlogController(db *gorm.DB) *blogController {
	return &blogController{db}
}

// Register binds the blog-related routes to a group
func (bc *blogController) Register(app *fiber.App) {
	// Blog CRUD routes
	app.Post("blogs/", bc.createBlog)                               // Create a new blog post
	app.Get("blogs/:blogId", bc.getSingleBlog)                      // Get a blog post by its ID
	app.Get("sessions/blogs", bc.getSessionBlogs)                   // Get all sessions blogs
	app.Get("blogs/", bc.getAllBlogsRandom)                         // Get all blogs
	app.Put("blogs/:blogId", bc.updateBlog)                         // Update a blog post
	app.Put("blogs/:blogId/likes", bc.likeBlog)                     // Like and/or unlike a blog post
	app.Delete("blogs/:blogId", bc.deleteBlog)                      // Delete a blog post
	app.Get("blogs/:blogId/follows/likes", bc.getLikesAndFollowers) // Get all likes and followed or followers that like a specific post

	// Comment-related routes
	app.Post("blogs/:blogId/comments", bc.addComment)        // Add a comment to a blog post
	app.Get("blogs/:blogId/comments", bc.getComments)        // Add a comment to a blog post
	app.Delete("comments/:commentId", bc.deleteComment)      // Delete a comment
	app.Put("comments/:commentId", bc.updateComment)         // Update a comment
	app.Put("comments/:commentId/likes", bc.likeComment)     // Like/unlike a comment
	app.Get("comments/:commentId/likes", bc.getCommentLikes) // Get likes for a comment
	app.Get("comments/:commentId/replies", bc.getReplies)    // Get replies to a comment

	// Reply-related routes
	app.Post("comments/:commentId/replies", bc.addReply) // Add reply to a comment

	// Follow/unfollow routes
	app.Post("users/follows", bc.followUnfollowUser) // Follow or unfollow a user

	// User-related routes
	app.Get("users/:userId/unfollowings", bc.getUsersNotFollowing) // Get users not followed by a specific user
	app.Get("users/:userId/followers", bc.getUserFollowers)        // Get a user's followers
	app.Get("users/:userId/followings", bc.getUserFollowings)      // Get a user's followings
	app.Get("blogs/users/:userId", bc.getUserBlogs)                // Get all user blogs

	// Users or bloggers
	app.Post("bloggers", bc.addBlogger)
	app.Get("bloggers", bc.getBloggers)
	app.Get("bloggers/:userId", bc.getBloggerByID)
	app.Delete("bloggers/:userId", bc.deleteBlogger)
}

// getAllBlogsRandom handles retrieving all blogs ordered randomly
func (bc *blogController) getAllBlogsRandom(c *fiber.Ctx) error {
	userId := c.Locals("userId")
	isAuthenticated := c.Locals("isAuthenticated").(bool)
	pageNumber, _ := strconv.Atoi(c.Query("pageNumber", "1"))
	numberOfRecords, _ := strconv.Atoi(c.Query("numberOfRecords", "20"))
	offset := (pageNumber - 1) * numberOfRecords

	var blogs []models.Blog
	if err := bc.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("profileImage", "fullName", "userId", "email", "verified")
	}).Order("RANDOM()").Limit(numberOfRecords).Offset(offset).Find(&blogs).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": fmt.Sprintf(" with Id %v does not exist", "")})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to fetch blogs"})
	}

	var returnBlogs []map[string]interface{} = []map[string]interface{}{}
	for _, blog := range blogs {
		var likesCount, sharesCount, commentsCount int64
		var liked, reposted bool

		// Count likes, shares, and comments
		bc.db.Model(&models.Like{}).Clauses(clause.Where{Exprs: []clause.Expression{
			clause.Eq{Column: "refId", Value: blog.BlogId},
		}}).Count(&likesCount)

		bc.db.Model(&models.Share{}).Clauses(clause.Where{Exprs: []clause.Expression{
			clause.Eq{Column: "refId", Value: blog.BlogId},
		}}).Count(&sharesCount)

		bc.db.Model(&models.Comment{}).Clauses(clause.Where{Exprs: []clause.Expression{
			clause.Eq{Column: "refId", Value: blog.BlogId},
		}}).Count(&commentsCount)

		// Check if the user has liked or reposted the blog
		if isAuthenticated {
			liked = bc.db.Model(&models.Like{}).Clauses(clause.Where{Exprs: []clause.Expression{
				clause.Eq{Column: "refId", Value: blog.BlogId},
				clause.Eq{Column: "userId", Value: userId},
			}}).Limit(1).Find(&models.Like{}).RowsAffected > 0

			reposted = bc.db.Model(&models.Share{}).Clauses(clause.Where{Exprs: []clause.Expression{
				clause.Eq{Column: "refId", Value: blog.BlogId},
				clause.Eq{Column: "userId", Value: userId},
			}}).Limit(1).Find(&models.Share{}).RowsAffected > 0
		}

		// Add blog data to the response
		returnBlogs = append(returnBlogs, fiber.Map{
			"blog":          blog,
			"liked":         liked,
			"reposted":      reposted,
			"likesCount":    likesCount,
			"repostsCount":  sharesCount,
			"commentsCount": commentsCount,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   returnBlogs,
	})
}

// getSingleBlog handles retrieving a single blog by its ID
func (bc *blogController) getSingleBlog(c *fiber.Ctx) error {
	blogId := c.Params("blogId")
	userId := c.Locals("userId")
	isAuthenticated := c.Locals("isAuthenticated").(bool)

	var blog models.Blog
	// Retrieve the blog with the associated user
	if err := bc.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("profileImage", "fullName", "userId", "email", "verified")
	}).Where(&models.Blog{BlogId: blogId}).First(&blog).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fmt.Sprintf("Blog with ID %s does not exist", blogId),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to fetch blogs"})

	}

	var likesCount, sharesCount, commentsCount int64
	var liked, reposted bool

	// Count likes, shares, and comments
	bc.db.Model(&models.Like{}).Where(&models.Like{RefId: blogId}).Count(&likesCount)
	bc.db.Model(&models.Share{}).Where(&models.Share{RefId: blogId}).Count(&sharesCount)
	bc.db.Model(&models.Comment{}).Where(&models.Comment{RefId: blogId}).Count(&commentsCount)

	// Check if the user has liked or reposted the blog
	if isAuthenticated {
		liked = bc.db.Model(&models.Like{}).Clauses(clause.Where{Exprs: []clause.Expression{
			clause.Eq{Column: "refId", Value: blogId},
			clause.Eq{Column: "userId", Value: userId},
		}}).Limit(1).Find(&models.Like{}).RowsAffected > 0

		reposted = bc.db.Model(&models.Share{}).Clauses(clause.Where{Exprs: []clause.Expression{
			clause.Eq{Column: "refId", Value: blogId},
			clause.Eq{Column: "userId", Value: userId},
		}}).Limit(1).Find(&models.Share{}).RowsAffected > 0
	}

	// Return blog data with additional metrics
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"blog":          blog,
			"liked":         liked,
			"reposted":      reposted,
			"likesCount":    likesCount,
			"repostsCount":  sharesCount,
			"commentsCount": commentsCount,
		},
	})
}

// getUserBlogs handles retrieving all blogs by a specific user
func (bc *blogController) getUserBlogs(c *fiber.Ctx) error {
	userId := c.Params("userId")
	isAuthenticated := c.Locals("isAuthenticated").(bool)

	pageNumber, _ := strconv.Atoi(c.Query("pageNumber", "1"))
	numberOfRecords, _ := strconv.Atoi(c.Query("numberOfRecords", "100"))
	offset := (pageNumber - 1) * numberOfRecords

	var blogs []models.Blog
	// Retrieve blogs by user ID with associated user information
	if err := bc.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("profileImage", "fullName", "userId", "email", "verified")
	}).Where(&models.Blog{UserId: userId}).Order(clause.OrderByColumn{Column: clause.Column{Name: "createdAt"}, Desc: true}).Limit(numberOfRecords).Offset(offset).Find(&blogs).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": fmt.Sprintf("User with ID %s does not exist", userId)})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to fetch blogs"})
	}

	var returnBlogs []map[string]interface{} = []map[string]interface{}{}
	for _, blog := range blogs {
		var likesCount, sharesCount, commentsCount int64
		var liked, reposted bool

		// Count likes, shares, and comments for each blog
		bc.db.Model(&models.Like{}).Where(&models.Like{RefId: blog.BlogId}).Count(&likesCount)
		bc.db.Model(&models.Share{}).Where(&models.Share{RefId: blog.BlogId}).Count(&sharesCount)
		bc.db.Model(&models.Comment{}).Where(&models.Comment{RefId: blog.BlogId}).Count(&commentsCount)

		// Check if the user has liked or reposted the blog
		if isAuthenticated {
			liked = bc.db.Model(&models.Like{}).Clauses(clause.Where{Exprs: []clause.Expression{
				clause.Eq{Column: "refId", Value: blog.BlogId},
				clause.Eq{Column: "userId", Value: userId},
			}}).Limit(1).Find(&models.Like{}).RowsAffected > 0

			reposted = bc.db.Model(&models.Share{}).Clauses(clause.Where{Exprs: []clause.Expression{
				clause.Eq{Column: "refId", Value: blog.BlogId},
				clause.Eq{Column: "userId", Value: userId},
			}}).Limit(1).Find(&models.Share{}).RowsAffected > 0
		}

		// Append blog data with additional metrics
		returnBlogs = append(returnBlogs, fiber.Map{
			"blog":          blog,
			"liked":         liked,
			"reposted":      reposted,
			"likesCount":    likesCount,
			"repostsCount":  sharesCount,
			"commentsCount": commentsCount,
		})
	}

	// Return all blogs data
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   returnBlogs,
	})
}

// getSessionBlogs handles retrieving blogs by the user's session
func (bc *blogController) getSessionBlogs(c *fiber.Ctx) error {
	userId := c.Locals("userId").(string)

	// Pagination parameters
	pageNumber, _ := strconv.Atoi(c.Query("pageNumber", "1"))
	numberOfRecords, _ := strconv.Atoi(c.Query("numberOfRecords", "20"))
	offset := (pageNumber - 1) * numberOfRecords

	// Retrieve follow IDs where the user is either a follower or being followed
	var followIds []interface{}
	if err := bc.db.Model(&models.Follow{}).Clauses(clause.Where{
		Exprs: []clause.Expression{
			clause.Or(
				clause.Eq{Column: "followerId", Value: userId},
				clause.Eq{Column: "followingId", Value: userId},
			),
		},
	}).Pluck("followingId", &followIds).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": fmt.Sprintf("User with ID %s does not exist", userId)})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve follow relationships"})
	}

	// Fetch blogs from followed users or the current user
	var blogs []models.Blog
	if err := bc.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("profileImage", "fullName", "userId", "email", "verified")
	}).Clauses(clause.Where{
		Exprs: []clause.Expression{
			clause.IN{Column: "userId", Values: append(followIds, userId)},
		},
	}).Order(clause.OrderByColumn{
		Column: clause.Column{Name: "createdAt"},
		Desc:   true,
	}).Limit(numberOfRecords).Offset(offset).Find(&blogs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to fetch blogs"})
	}

	// Prepare the response with additional metadata for each blog
	var returnBlogs []map[string]interface{} = []map[string]interface{}{}
	for _, blog := range blogs {
		var likesCount, sharesCount, commentsCount int64
		var liked, reposted bool

		// Count likes, shares, and comments for the blog
		bc.db.Model(&models.Like{}).Where(&models.Like{RefId: blog.BlogId}).Count(&likesCount)
		bc.db.Model(&models.Share{}).Where(&models.Share{RefId: blog.BlogId}).Count(&sharesCount)
		bc.db.Model(&models.Comment{}).Where(&models.Comment{RefId: blog.BlogId}).Count(&commentsCount)

		// Check if the current user liked or reposted the blog
		liked = bc.db.Model(&models.Like{}).Clauses(clause.Where{
			Exprs: []clause.Expression{
				clause.Eq{Column: "refId", Value: blog.BlogId},
				clause.Eq{Column: "userId", Value: userId},
			},
		}).Limit(1).Find(&models.Like{}).RowsAffected > 0

		reposted = bc.db.Model(&models.Share{}).Clauses(clause.Where{
			Exprs: []clause.Expression{
				clause.Eq{Column: "refId", Value: blog.BlogId},
				clause.Eq{Column: "userId", Value: userId},
			},
		}).Limit(1).Find(&models.Share{}).RowsAffected > 0

		// Add the blog details and metadata to the response
		returnBlogs = append(returnBlogs, fiber.Map{
			"blog":          blog,
			"liked":         liked,
			"reposted":      reposted,
			"likesCount":    likesCount,
			"repostsCount":  sharesCount,
			"commentsCount": commentsCount,
		})
	}

	// Return the response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   returnBlogs,
	})
}

func (bc *blogController) createBlog(c *fiber.Ctx) error {
	var blogPayload struct {
		UserId            string   `json:"userId"`
		Slug              string   `json:"slug"`
		Title             string   `json:"title"`
		URL               string   `json:"url"`
		ExternalLink      string   `json:"externalLink"`
		ExternalLinkTitle string   `json:"externalLinkTitle"`
		Text              string   `json:"text"`
		Images            []string `json:"images"`
		Video             string   `json:"video"`
		Reposted          bool     `json:"reposted"`
		FromBlogId        string   `json:"fromBlogId,omitempty"`
	}

	if err := c.BodyParser(&blogPayload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userId := c.Locals("userId").(string)
	slug := uuid.New().String()

	modifiedBlogPayload := models.Blog{
		Title:             blogPayload.Title,
		Text:              blogPayload.Text,
		Slug:              slug,
		UserId:            userId,
		ExternalLink:      blogPayload.ExternalLink,
		ExternalLinkTitle: blogPayload.ExternalLinkTitle,
		Images:            blogPayload.Images,
		Video:             blogPayload.Video,
	}

	// Use pq.Array to handle the Images slice
	if err := bc.db.Model(&models.Blog{}).Create(map[string]interface{}{
		"Title":             modifiedBlogPayload.Title,
		"Text":              modifiedBlogPayload.Text,
		"Slug":              modifiedBlogPayload.Slug,
		"UserId":            modifiedBlogPayload.UserId,
		"ExternalLink":      modifiedBlogPayload.ExternalLink,
		"ExternalLinkTitle": modifiedBlogPayload.ExternalLinkTitle,
		"Images":            modifiedBlogPayload.Images,
		"Video":             modifiedBlogPayload.Video,
		"createdAt":         time.Now(),
		"updatedAt":         time.Now(),
	}).Preload("User").First(&modifiedBlogPayload).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create blog"})
	}

	if blogPayload.Reposted && blogPayload.FromBlogId != "" {
		if err := bc.db.Create(&models.Share{UserId: userId, RefId: blogPayload.FromBlogId}).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to record repost"})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Successfully added a blog",
		"data":    modifiedBlogPayload,
	})
}

// updateBlog handles updating a blog
func (bc *blogController) updateBlog(c *fiber.Ctx) error {

	var blogPayload struct {
		UserId            string   `json:"userId"`
		Slug              string   `json:"slug"`
		Title             string   `json:"title"`
		URL               string   `json:"url"`
		ExternalLink      string   `json:"externalLink"`
		ExternalLinkTitle string   `json:"externalLinkTitle"`
		Text              string   `json:"text"`
		Images            []string `json:"images"`
		Video             string   `json:"video"`
		Reposted          bool     `json:"reposted"`
		FromBlogId        string   `json:"fromBlogId,omitempty"`
	}

	blogId := c.Params("blogId")

	// Parse the request body
	if err := c.BodyParser(&blogPayload); err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Retrieve the blog by blogId using Where for a single condition
	var blog models.Blog
	if err := bc.db.Where(&models.Blog{BlogId: blogId}).First(&blog).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": fmt.Sprintf("Blog with ID %s does not exist", blogId)})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Couldn't update blog"})
	}

	// Update the blog's data using Model and Updates
	if err := bc.db.Model(&blog).Updates(blogPayload).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update blog"})
	}

	// Return updated blog data
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"status": "success",
		"data":   blog,
	})
}

// deleteBlog handles deleting a blog and related entities
func (bc *blogController) deleteBlog(c *fiber.Ctx) error {
	blogId := c.Params("blogId")

	var blog models.Blog
	// Retrieve blog by blogId using Where for a single condition
	if err := bc.db.Where(&models.Blog{BlogId: blogId}).First(&blog).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": fmt.Sprintf("Blog with ID %s does not exist", blogId)})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Couldn't delete blog"})
	}

	// Perform cascading deletions within a transaction
	if err := bc.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(&models.Comment{RefId: blogId}).Delete(&models.Comment{}).Error; err != nil {
			return err
		}
		if err := tx.Where(&models.Like{RefId: blogId}).Delete(&models.Like{}).Error; err != nil {
			return err
		}
		if err := tx.Where(&models.Share{RefId: blogId}).Delete(&models.Share{}).Error; err != nil {
			return err
		}
		if err := tx.Where(&models.Blog{BlogId: blogId}).Delete(&models.Blog{}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete blog"})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Successfully deleted a blog",
	})
}

// likeBlog handles liking or unliking a blog
func (bc *blogController) likeBlog(c *fiber.Ctx) error {
	userId := c.Locals("userId").(string)
	blogId := c.Params("blogId")

	var like models.Like
	// Check if the user has already liked the blog using Where for a compound condition
	if err := bc.db.Where(&models.Like{UserId: userId, RefId: blogId}).First(&like).Error; err == nil {
		// Unlike logic
		if err := bc.db.Delete(&like).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to unlike the post"})
		}

		// Get the updated likes count
		var likesCount int64
		bc.db.Model(&models.Like{}).Where(&models.Like{RefId: blogId}).Count(&likesCount)

		// Return response
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":     "success",
			"liked":      false,
			"likesCount": likesCount,
		})
	}

	// Like logic
	newLike := models.Like{
		UserId:    userId,
		RefId:     blogId,
		CreatedAt: time.Now(),
	}

	if err := bc.db.Create(&newLike).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to like the post"})
	}

	// Get the updated likes count
	var likesCount int64
	bc.db.Model(&models.Like{}).Where(&models.Like{RefId: blogId}).Count(&likesCount)

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":     "success",
		"liked":      true,
		"likesCount": likesCount,
	})
}

// GetReplies fetches all replies for a specific comment with pagination
func (bc *blogController) getComments(c *fiber.Ctx) error {
	blogId := c.Params("blogId")
	pageNumber, _ := strconv.Atoi(c.Query("pageNumber", "1"))
	numberOfRecords, _ := strconv.Atoi(c.Query("numberOfRecords", "100"))
	offset := (pageNumber - 1) * numberOfRecords

	var replies []models.Comment = []models.Comment{}
	// Use Where model pattern for filtering
	if err := bc.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("profileImage", "userId", "fullName", "email", "verified")
	}).Where(&models.Comment{RefId: blogId}).Order(clause.OrderByColumn{Column: clause.Column{Name: "createdAt"}, Desc: true}).Limit(numberOfRecords).Offset(offset).Find(&replies).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": fmt.Sprintf("Blog with ID %s does not exist", blogId)})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Couldn't update blog"})
	}

	if len(replies) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": replies})
	}

	// Add repliesCount and likesCount to each reply
	enrichedReplies := make([]fiber.Map, len(replies))
	for i, reply := range replies {
		var repliesCount int64
		// Use Where model pattern for filtering
		bc.db.Model(&models.Comment{}).Where(&models.Comment{RefId: reply.CommentId}).Count(&repliesCount)

		var likesCount int64
		// Use Where model pattern for filtering
		bc.db.Model(&models.Like{}).Where(&models.Like{RefId: reply.CommentId}).Count(&likesCount)

		enrichedReplies[i] = fiber.Map{
			"comment":       reply,
			"commentsCount": repliesCount,
			"likesCount":    likesCount,
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   enrichedReplies,
	})
}

// addComment handles adding a comment to a blog
func (bc *blogController) addComment(c *fiber.Ctx) error {
	var payload struct {
		Content string `json:"content"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userId := c.Locals("userId").(string)
	blogId := c.Params("blogId")

	// Create the comment
	comment := models.Comment{
		UserId:  userId,
		Content: payload.Content,
		RefId:   blogId,
	}

	if err := bc.db.Create(&comment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add comment"})
	}

	// Fetch replies count for the comment
	var repliesCount int64
	bc.db.Model(&models.Comment{}).Where(&models.Comment{RefId: comment.CommentId}).Count(&repliesCount)

	// Fetch likes count for the comment
	var likesCount int64
	bc.db.Model(&models.Like{}).Where(&models.Like{RefId: comment.CommentId}).Count(&likesCount)

	// Fetch user details
	var user models.User
	if err := bc.db.Select("profileImage", "userId", "fullName", "email", "verified").Where(&models.User{UserId: userId}).First(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch user details"})
	}

	// Check if the comment is liked by the user
	isLiked := bc.db.Model(&models.Like{}).Where(&models.Like{RefId: comment.CommentId, UserId: userId}).RowsAffected > 0

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("Successfully added a comment to blogId = %s", blogId),
		"data": fiber.Map{
			"comment":      comment,
			"repliesCount": repliesCount,
			"likesCount":   likesCount,
			"User":         user,
			"liked":        isLiked,
		},
	})
}

// deleteComment handles deleting a comment and related entities
func (bc *blogController) deleteComment(c *fiber.Ctx) error {
	commentId := c.Params("commentId")

	// Find the comment using the Where clause for a single condition
	var comment models.Comment
	if err := bc.db.Where(&models.Comment{CommentId: commentId}).First(&comment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprintf("Comment with ID %s does not exist", commentId),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Couldn't delete comment"})

	}

	// Perform cascading deletions within a transaction
	if err := bc.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(&models.Comment{RefId: commentId}).Delete(&models.Comment{}).Error; err != nil {
			return err
		}
		if err := tx.Where(&models.Like{RefId: commentId}).Delete(&models.Like{}).Error; err != nil {
			return err
		}
		return tx.Where(&models.Comment{CommentId: commentId}).Delete(&comment).Error
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete comment"})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Successfully deleted the comment",
	})
}

// updateComment handles updating a comment
func (bc *blogController) updateComment(c *fiber.Ctx) error {
	commentId := c.Params("commentId")
	var payload struct {
		Content string `json:"content"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Find the comment using the Where clause
	var comment models.Comment
	if err := bc.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("profileImage", "userId", "fullName", "email", "verified")
	}).Where(&models.Comment{CommentId: commentId}).First(&comment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprintf("Comment with ID %s does not exist", commentId),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Couldn't update comment"})
	}

	// Update the comment content
	if err := bc.db.Model(&comment).Update("content", payload.Content).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update comment"})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"status":  "success",
		"message": "Successfully updated the comment",
		"data":    comment,
	})
}

// likeComment handles liking or unliking a comment
func (bc *blogController) likeComment(c *fiber.Ctx) error {
	userId := c.Locals("userId").(string)
	commentId := c.Params("commentId")

	var like models.Like
	// Check if the user has already liked the comment
	if err := bc.db.Where(&models.Like{UserId: userId, RefId: commentId}).First(&like).Error; err == nil {
		// Unlike logic
		if err := bc.db.Delete(&like).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to unlike the comment"})
		}

		// Get the updated likes count
		var likesCount int64
		bc.db.Model(&models.Like{}).Where(&models.Like{RefId: commentId}).Count(&likesCount)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":     "success",
			"liked":      false,
			"likesCount": likesCount,
		})
	}

	// Like logic
	newLike := models.Like{
		UserId:    userId,
		RefId:     commentId,
		CreatedAt: time.Now(),
	}

	// Create a new like
	if err := bc.db.Create(&newLike).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to like the comment"})
	}

	// Get the updated likes count
	var likesCount int64
	bc.db.Model(&models.Like{}).Where(&models.Like{RefId: commentId}).Count(&likesCount)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":     "success",
		"liked":      true,
		"likesCount": likesCount,
	})
}

// GetCommentLikes fetches all likes for a specific comment
func (bc *blogController) getCommentLikes(c *fiber.Ctx) error {
	commentId := c.Params("commentId")

	var likes []models.Like = []models.Like{}
	// Use Where model pattern for filtering
	if err := bc.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("profileImage", "userId", "fullName", "email", "verified")
	}).Where(&models.Like{RefId: commentId}).Find(&likes).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprintf("Comment with ID %s does not exist", commentId),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch likes"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   likes,
	})
}

// GetReplies fetches all replies for a specific comment with pagination
func (bc *blogController) getReplies(c *fiber.Ctx) error {
	commentId := c.Params("commentId")
	pageNumber, _ := strconv.Atoi(c.Query("pageNumber", "1"))
	numberOfRecords, _ := strconv.Atoi(c.Query("numberOfRecords", "100"))
	offset := (pageNumber - 1) * numberOfRecords

	var replies []models.Comment = []models.Comment{}
	// Use Where model pattern for filtering
	if err := bc.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("profileImage", "userId", "fullName", "email", "verified")
	}).Where(&models.Comment{RefId: commentId}).Order(clause.OrderByColumn{Column: clause.Column{Name: "createdAt"}, Desc: true}).Limit(numberOfRecords).Offset(offset).Find(&replies).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprintf("Comment with ID %s does not exist", commentId),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Couldn't delete comment"})
	}

	if len(replies) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": replies})
	}

	// Add repliesCount and likesCount to each reply
	enrichedReplies := make([]fiber.Map, len(replies))
	for i, reply := range replies {
		var repliesCount int64
		// Use Where model pattern for filtering
		bc.db.Model(&models.Comment{}).Where(&models.Comment{RefId: reply.CommentId}).Count(&repliesCount)

		var likesCount int64
		// Use Where model pattern for filtering
		bc.db.Model(&models.Like{}).Where(&models.Like{RefId: reply.CommentId}).Count(&likesCount)

		enrichedReplies[i] = fiber.Map{
			"comment":      reply,
			"repliesCount": repliesCount,
			"likesCount":   likesCount,
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   enrichedReplies,
	})
}

// AddReply adds a reply to a specific comment
func (bc *blogController) addReply(c *fiber.Ctx) error {
	commentId := c.Params("commentId")
	userId := c.Locals("userId").(string)

	var payload struct {
		Content string `json:"content"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Create the reply
	reply := models.Comment{
		RefId:   commentId,
		UserId:  userId,
		Content: payload.Content,
	}

	// Use model-based Create approach
	if err := bc.db.Create(&reply).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": fmt.Sprintf("Comment with ID %v does not exist", commentId)})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Couldn't delete comment"})

	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("Successfully added a reply to commentId = %s", commentId),
		"data":    reply,
	})
}

// FollowUnfollowUser allows a user to follow or unfollow another user
func (bc *blogController) followUnfollowUser(c *fiber.Ctx) error {
	var payload struct {
		FollowerId  string `json:"followerId"`
		FollowingId string `json:"followingId"`
	}

	// Parse the request body to get followerId and followingId
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	followerId := payload.FollowerId
	followingId := payload.FollowingId

	// Check if the user is already following the target user
	var follow models.Follow
	err := bc.db.Where(&models.Follow{FollowerId: followerId, FollowingId: followingId}).First(&follow).Error

	if err == nil {
		// If a follow record exists, unfollow the user
		if err := bc.db.Delete(&follow).Error; err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "Failed to unfollow"})
		}

		// Optionally, check the number of followers and update verification status if needed
		var followerCount int64
		bc.db.Model(&models.Follow{}).Where(&models.Follow{FollowingId: followingId}).Count(&followerCount)

		// Verification logic when the follower count reaches 30
		if followerCount == 30 {
			if err := bc.db.Model(&models.User{}).Where(&models.User{UserId: followingId}).Update("verified", true).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to verify user"})
			}
		}

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"message": "Unfollowed successfully",
			"data": fiber.Map{
				"followed": false,
			},
		})
	}

	// If no follow record exists, follow the user
	follow = models.Follow{
		FollowerId:  followerId,
		FollowingId: followingId,
	}

	// Use model-based Create approach
	if err := bc.db.Create(&follow).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to follow"})
	}

	// Check if the new following user reaches the follower limit and update verification
	var followerCount int64
	bc.db.Model(&models.Follow{}).Where(&models.Follow{FollowingId: followingId}).Count(&followerCount)

	// Verification logic when the follower count reaches 30
	if followerCount == 30 {
		if err := bc.db.Model(&models.User{}).Where(&models.User{UserId: followingId}).Update("verified", true).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to verify user"})
		}
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Followed successfully",
		"data": fiber.Map{
			"followed": true,
		},
	})
}

// GetUsersNotFollowing retrieves users who are not followed by a specific user
func (bc *blogController) getUsersNotFollowing(c *fiber.Ctx) error {
	userId := c.Params("userId")
	pageNumber, _ := strconv.Atoi(c.Query("pageNumber", "1"))
	numberOfRecords, _ := strconv.Atoi(c.Query("numberOfRecords", "100"))
	offset := (pageNumber - 1) * numberOfRecords

	var ids []interface{}
	// Get all followingIds where userId is the follower
	bc.db.Model(&models.Follow{}).Where(&models.Follow{FollowerId: userId}).Pluck("followingId", &ids)

	// Fetch users who are not followed by the current user
	var users []models.User = []models.User{}
	err := bc.db.Clauses(clause.Where{Exprs: []clause.Expression{clause.Not(clause.IN{Column: "userId", Values: ids})}}).
		Limit(numberOfRecords).Offset(offset).
		Find(&users).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch users"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   users,
	})
}

// GetUserFollowers retrieves the followers of a specific user
func (bc *blogController) getUserFollowers(c *fiber.Ctx) error {
	userId := c.Params("userId")
	pageNumber, _ := strconv.Atoi(c.Query("pageNumber", "1"))
	numberOfRecords, _ := strconv.Atoi(c.Query("numberOfRecords", "100"))
	offset := (pageNumber - 1) * numberOfRecords

	var followerIds []interface{}
	// Get all followerIds where userId is the followingId
	bc.db.Model(&models.Follow{}).Where(&models.Follow{FollowingId: userId}).Pluck("followerId", &followerIds)

	var users []models.User = []models.User{}
	err := bc.db.Clauses(clause.Where{Exprs: []clause.Expression{clause.IN{Column: "userId", Values: followerIds}}}).
		Limit(numberOfRecords).Offset(offset).
		Find(&users).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch followers"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   users,
	})
}

// Get a user's followings
func (bc *blogController) getUserFollowings(c *fiber.Ctx) error {
	userId := c.Params("userId")
	// activeUserId := c.Locals("userId").(string)

	var followingIds []interface{}
	bc.db.Model(&models.Follow{}).Where(&models.Follow{FollowerId: userId}).Pluck("followingId", &followingIds)

	var users []models.User = []models.User{}
	err := bc.db.Clauses(clause.Where{Exprs: []clause.Expression{clause.IN{Column: "userId", Values: followingIds}}}).
		Find(&users).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch followings"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   users,
	})
}

// Get all likes and followed or followers that like a specific blog
func (bc *blogController) getLikesAndFollowers(c *fiber.Ctx) error {
	blogId := c.Params("blogId")
	currentUserId := c.Locals("userId").(string)
	var userIds []string
	var likes []models.Like
	if err := bc.db.Where(&models.Like{RefId: blogId}).Find(&likes).Pluck("userId", &userIds).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch likes"})
	}

	// for _, like := range likes {
	// 	userIds = append(userIds, like.UserId)
	// }

	var userFollowings []models.Follow
	bc.db.Clauses(clause.Where{
		Exprs: []clause.Expression{
			clause.Or(
				clause.Eq{Column: "followerId", Value: currentUserId},
				clause.Eq{Column: "followingId", Value: currentUserId},
			),
		},
	}).Find(&userFollowings)

	var commonUserIds []interface{}
	for _, userFollowing := range userFollowings {
		if contains(userIds, userFollowing.FollowerId) || contains(userIds, userFollowing.FollowingId) {
			commonUserIds = append(commonUserIds, userFollowing.FollowerId)
		}
	}

	var users []models.User = []models.User{}
	err := bc.db.Clauses(clause.Where{Exprs: []clause.Expression{clause.IN{Column: "userId", Values: commonUserIds}}}).Find(&users).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch users"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			// "likes":        likes,
			"sessionUsers": users,
		},
	})
}

// Utility function
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// ////////////////////////////////////////// Users section ////////////////////////////////
func (bc *blogController) addBlogger(c *fiber.Ctx) error {
	var user models.User

	// Parse the incoming request body into the user struct
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	// Hash the user's password before saving
	hashedPassword, err := utils.HashData(user.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to hash password"})
	}

	// Replace the plain password with the hashed password
	user.Password = hashedPassword

	// Save the user in the database (assuming `Create` is a GORM method)
	if err := bc.db.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to add user"})
	}

	// Respond with success
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Successfully added a user",
		"data":    user,
	})
}

func (bc *blogController) getBloggers(c *fiber.Ctx) error {
	pageNumber, _ := strconv.Atoi(c.Query("pageNumber", "1"))
	numberOfRecords, _ := strconv.Atoi(c.Query("numberOfRecords", "100"))
	offset := (pageNumber - 1) * numberOfRecords

	var users []models.User
	if err := bc.db.Limit(numberOfRecords).Offset(offset).Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch users"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   users,
	})
}

func (bc *blogController) getBloggerByID(c *fiber.Ctx) error {
	userId := c.Params("userId")
	var user models.User
	if err := bc.db.Model(&user).Where(&models.User{UserId: userId}).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": fmt.Sprintf("User with ID %s not found", userId)})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   user,
	})
}

func (bc *blogController) deleteBlogger(c *fiber.Ctx) error {
	userId := c.Params("userId")
	var user models.User
	if err := bc.db.Model(&user).Where(&models.User{UserId: userId}).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("User with ID %s does not exist", userId),
		})
	}

	if err := bc.db.Delete(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to delete user"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Successfully deleted the user",
	})
}
