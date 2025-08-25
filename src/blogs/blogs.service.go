package blogs

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/epsierra/phinex-blog-api/src/models"
	"github.com/epsierra/phinex-blog-api/src/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BlogsService handles blog-related operations
type BlogsService struct {
	db     *gorm.DB
	logger *log.Logger
}

// NewBlogsService creates a new BlogsService instance
func NewBlogsService(db *gorm.DB) *BlogsService {
	return &BlogsService{
		db:     db,
		logger: log.New(os.Stderr, "blogs-service: ", log.LstdFlags),
	}
}

// FindAll retrieves all blogs ordered by created_at descending
func (s *BlogsService) FindAll(currentUser models.ICurrentUser, page, limit int) (models.PaginatedResponse, error) {
	var totalItems int64
	s.db.Model(&models.Blog{}).Count(&totalItems)

	offset := (page - 1) * limit
	var blogs []models.Blog
	err := s.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("profile_image", "full_name", "user_id", "email", "verified")
	}).Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).Limit(limit).Offset(offset).Find(&blogs).Error
	if err != nil {
		s.logger.Printf("Error fetching blogs: %v", err)
		return models.PaginatedResponse{Data: []BlogWithMeta{}}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Unable to fetch blogs"}
	}

	blogsWithMeta, err := s.enrichBlogs(blogs, currentUser)
	if err != nil {
		return models.PaginatedResponse{Data: []BlogWithMeta{}}, err
	}

	totalPages := (totalItems + int64(limit) - 1) / int64(limit)

	return models.PaginatedResponse{
		Data: blogsWithMeta,
		Metadata: models.PaginationMetadata{
			CurrentPage:     int64(page),
			ItemsPerPage:    int64(limit),
			TotalItems:      totalItems,
			TotalPages:      totalPages,
			HasNextPage:     int64(page) < totalPages,
			HasPreviousPage: int64(page) > 1,
		},
	}, nil
}

// FindOne retrieves a single blog by ID
func (s *BlogsService) FindOne(blogId string, currentUser models.ICurrentUser) (BlogWithMeta, error) {
	var blog models.Blog
	err := s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("profile_image", "full_name", "user_id", "email", "verified")
		}).Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "blog_id", Value: blogId}}}).
			First(&blog).Error
		if err != nil {
			return err
		}

		// Increment views_count for the blog
		if err := tx.Model(&models.Blog{}).Where("blog_id = ?", blogId).Update("views_count", gorm.Expr("views_count + ?", 1)).Error; err != nil {
			return err
		}

		// Create a new View entry
		view := models.View{
			ViewId:    utils.GenerateID(),
			RefId:     blogId,
			UserId:    currentUser.UserId,
			CreatedAt: time.Now().UTC(),
			CreatedBy: currentUser.FullName,
			UpdatedAt: time.Now().UTC(),
			UpdatedBy: currentUser.FullName,
		}
		if err := tx.Create(&view).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return BlogWithMeta{}, &fiber.Error{Code: fiber.StatusNotFound, Message: fmt.Sprintf("Blog with ID %s does not exist", blogId)}
		}
		s.logger.Printf("Error fetching blog: %v", err)
		return BlogWithMeta{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Unable to fetch blog"}
	}

	blogsWithMeta, err := s.enrichBlogs([]models.Blog{blog}, currentUser)
	if err != nil {
		return BlogWithMeta{}, err
	}
	return blogsWithMeta[0], nil
}

// FindFollowingBlogs retrieves blogs from followed users or the current user
func (s *BlogsService) FindFollowingBlogs(currentUser models.ICurrentUser, page, limit int) (models.PaginatedResponse, error) {
	var totalItems int64
	s.db.Model(&models.Blog{}).Count(&totalItems)

	offset := (page - 1) * limit
	var followIds []interface{}
	err := s.db.Model(&models.Follow{}).Clauses(clause.Where{
		Exprs: []clause.Expression{
			clause.Or(
				clause.Eq{Column: "follower_id", Value: currentUser.UserId},
				clause.Eq{Column: "following_id", Value: currentUser.UserId},
			),
		},
	}).Pluck("following_id", &followIds).Error
	if err != nil {
		s.logger.Printf("Error fetching follow IDs: %v", err)
		return models.PaginatedResponse{Data: []BlogWithMeta{}}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to retrieve follow relationships"}
	}

	var blogs []models.Blog
	err = s.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("profile_image", "full_name", "user_id", "email", "verified").Preload("UserRoles.Role")
	}).Clauses(clause.Where{
		Exprs: []clause.Expression{clause.IN{Column: "user_id", Values: append(followIds, currentUser.UserId)}},
	}).Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).
		Limit(limit).Offset(offset).Find(&blogs).Error
	if err != nil {
		s.logger.Printf("Error fetching session blogs: %v", err)
		return models.PaginatedResponse{Data: []BlogWithMeta{}}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Unable to fetch blogs"}
	}

	blogsWithMeta, err := s.enrichBlogs(blogs, currentUser)
	if err != nil {
		return models.PaginatedResponse{Data: []BlogWithMeta{}}, err
	}

	totalPages := (totalItems + int64(limit) - 1) / int64(limit)

	return models.PaginatedResponse{
		Data: blogsWithMeta,
		Metadata: models.PaginationMetadata{
			CurrentPage:     int64(page),
			ItemsPerPage:    int64(limit),
			TotalItems:      totalItems,
			TotalPages:      totalPages,
			HasNextPage:     int64(page) < totalPages,
			HasPreviousPage: int64(page) > 1,
		},
	}, nil
}

// FindUserBlogs retrieves blogs by a specific user
func (s *BlogsService) FindUserBlogs(userId string, currentUser models.ICurrentUser, page, limit int) (models.PaginatedResponse, error) {
	var totalItems int64
	s.db.Model(&models.Blog{}).Where("user_id = ?", userId).Count(&totalItems)

	offset := (page - 1) * limit
	var blogs []models.Blog
	err := s.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("profile_image", "full_name", "user_id", "email", "verified").Preload("UserRoles.Role")
	}).Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "user_id", Value: userId}}}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).
		Limit(limit).Offset(offset).Find(&blogs).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.PaginatedResponse{Data: []BlogWithMeta{}}, &fiber.Error{Code: fiber.StatusNotFound, Message: fmt.Sprintf("User with ID %s does not exist", userId)}
		}
		s.logger.Printf("Error fetching user blogs: %v", err)
		return models.PaginatedResponse{Data: []BlogWithMeta{}}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Unable to fetch blogs"}
	}

	blogsWithMeta, err := s.enrichBlogs(blogs, currentUser)
	if err != nil {
		return models.PaginatedResponse{Data: []BlogWithMeta{}}, err
	}

	totalPages := (totalItems + int64(limit) - 1) / int64(limit)

	return models.PaginatedResponse{
		Data: blogsWithMeta,
		Metadata: models.PaginationMetadata{
			CurrentPage:     int64(page),
			ItemsPerPage:    int64(limit),
			TotalItems:      totalItems,
			TotalPages:      totalPages,
			HasNextPage:     int64(page) < totalPages,
			HasPreviousPage: int64(page) > 1,
		},
	}, nil
}

// Create creates a new blog
func (s *BlogsService) Create(dto CreateBlogDto, currentUser models.ICurrentUser) (MutationResponse, error) {

	var blog models.Blog
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Marshal Images to JSON
		imagesJSON, err := json.Marshal(dto.Images)
		if err != nil {
			return fmt.Errorf("failed to marshal images: %w", err)
		}

		blogId := utils.GenerateID()

		blog = models.Blog{
			BlogId:            blogId,
			UserId:            currentUser.UserId,
			Title:             dto.Title,
			ExternalLink:      dto.ExternalLink,
			ExternalLinkTitle: dto.ExternalLinkTitle,
			Text:              dto.Text,
			Images:            imagesJSON,
			Video:             dto.Video,
			Audio:             dto.Audio,
			IsReel:            dto.Video != "",
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
			CreatedBy:         currentUser.FullName,
			UpdatedBy:         currentUser.FullName,
		}
		if err := tx.Create(&blog).First(&blog).Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("profile_image", "full_name", "user_id", "email", "verified")
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.UsersStats{}).Where("user_id = ?", currentUser.UserId).Update("total_posts", gorm.Expr("total_posts + 1")).Error; err != nil {
			return err
		}

		if dto.Pinned && dto.RepostedFromBlogId == "" {
			pinnedBlog := models.PinnedBlog{
				PinnedBlogId: utils.GenerateID(),
				BlogId:       blogId,
				StartDate:    time.Now().UTC(),
				EndDate:      time.Now().UTC().AddDate(0, 0, dto.PinnedNumerOfDays),
				UserId:       currentUser.UserId,
				CreatedAt:    time.Now().UTC(),
				UpdatedAt:    time.Now().UTC(),
				CreatedBy:    currentUser.FullName,
				UpdatedBy:    currentUser.FullName,
			}

			if err := tx.Create(&pinnedBlog).Error; err != nil {
				return err
			}

			// Make payment placeholder
		}
		if dto.RepostedFromBlogId != "" {

			// Increment sharesCount on the associated blog
			if err := tx.Model(&models.Blog{}).Where("blog_id = ?", dto.RepostedFromBlogId).Update("shares_count", gorm.Expr("shares_count + ?", 1)).Error; err != nil {
				return err
			}
			share := models.Share{
				ShareId:   utils.GenerateID(),
				UserId:    currentUser.UserId,
				RefId:     dto.RepostedFromBlogId,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				CreatedBy: currentUser.FullName,
				UpdatedBy: currentUser.FullName,
			}
			return tx.Create(&share).Error
		}
		return nil
	})
	if err != nil {
		s.logger.Printf("Error creating blog: %v", err)
		return MutationResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to create blog"}
	}

	return MutationResponse{
		Message: "Blog created successfully",
		Data:    blog,
	}, nil
}

// Update updates a blog
func (s *BlogsService) Update(blogId string, dto UpdateBlogDto, currentUser models.ICurrentUser) (MutationResponse, error) {
	var blog models.Blog
	err := s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "blog_id", Value: blogId}}}).
			First(&blog).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &fiber.Error{Code: fiber.StatusNotFound, Message: fmt.Sprintf("Blog with ID %s does not exist", blogId)}
			}
			return err
		}

		updateData := map[string]interface{}{
			"updated_at": time.Now().UTC(),
			"updated_by": currentUser.FullName,
		}
		if dto.Title != "" {
			updateData["title"] = dto.Title
		}

		if dto.ExternalLink != "" {
			updateData["external_link"] = dto.ExternalLink
		}
		if dto.ExternalLinkTitle != "" {
			updateData["external_link_title"] = dto.ExternalLinkTitle
		}
		if dto.Text != "" {
			updateData["text"] = dto.Text
		}
		if len(dto.Images) > 0 {
			// Marshal Images to JSON
			imagesJSON, err := json.Marshal(dto.Images)
			if err != nil {
				return fmt.Errorf("failed to marshal images: %w", err)
			}
			updateData["images"] = imagesJSON
		}
		if dto.Video != "" {
			updateData["video"] = dto.Video
			updateData["is_reel"] = true
		}

		if dto.Audio != "" {
			updateData["audio"] = dto.Audio
		}
		return tx.Model(&blog).Updates(updateData).Error
	})
	if err != nil {
		s.logger.Printf("Error updating blog: %v", err)
		if fiberErr, ok := err.(*fiber.Error); ok {
			return MutationResponse{}, fiberErr
		}
		return MutationResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to update blog"}
	}

	return MutationResponse{
		Message: "Blog updated successfully",
		Data:    blog,
	}, nil
}

// Delete deletes a blog
func (s *BlogsService) Delete(blogId string, currentUser models.ICurrentUser) (map[string]string, error) {

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var blog models.Blog
		err := tx.Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "blog_id", Value: blogId}}}).
			First(&blog).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &fiber.Error{Code: fiber.StatusNotFound, Message: fmt.Sprintf("Blog with ID %s does not exist", blogId)}
			}
			return err
		}

		if err := tx.Model(&models.PinnedBlog{}).Where("blog_id = ?", blogId).Delete(&models.PinnedBlog{}).Error; err != nil { // Unpin if pinned
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Do nothing if not found
			} else {
				return err
			}
		}

		if err := tx.Model(&models.UsersStats{}).Where("user_id = ?", blog.UserId).Update("total_posts", gorm.Expr("total_posts - 1")).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "ref_id", Value: blogId}}}).
			Delete(&models.Comment{}).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Do nothing if not found
			} else {
				return err
			}
		}
		if err := tx.Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "ref_id", Value: blogId}}}).
			Delete(&models.Like{}).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Do nothing if not found
			} else {
				return err
			}
		}
		if err := tx.Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "ref_id", Value: blogId}}}).
			Delete(&models.Share{}).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Do nothing if not found
			} else {
				return err
			}
		}
		return tx.Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "blog_id", Value: blogId}}}).
			Delete(&blog).Error
	})
	if err != nil {
		s.logger.Printf("Error deleting blog: %v", err)
		if fiberErr, ok := err.(*fiber.Error); ok {
			return nil, fiberErr
		}
		return nil, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to delete blog"}
	}

	return map[string]string{"message": "Blog deleted successfully"}, nil
}

// LikeBlog handles liking or unliking a blog
func (s *BlogsService) LikeBlog(blogId string, currentUser models.ICurrentUser) (LikeResponse, error) {
	if !currentUser.IsAuthenticated {
		return LikeResponse{}, &fiber.Error{Code: fiber.StatusUnauthorized, Message: "Unauthorized"}
	}

	var like models.Like
	var blog models.Blog
	err := s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Where{
			Exprs: []clause.Expression{
				clause.Eq{Column: "user_id", Value: currentUser.UserId},
				clause.Eq{Column: "ref_id", Value: blogId},
			},
		}).First(&like).Error
		if err == nil {
			// Unlike
			if err := tx.Delete(&like).Error; err != nil {
				return err
			}
			// Decrement likesCount on the associated blog
			if err := tx.Model(&models.Blog{}).Where("blog_id = ?", blogId).Update("likes_count", gorm.Expr("likes_count - ?", 1)).Error; err != nil {
				return err
			}

			return tx.Model(&models.UsersStats{}).Where("user_id = ?", currentUser.UserId).Update("total_likes", gorm.Expr("total_likes - 1")).Error
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// Like
		newLike := models.Like{
			LikeId:    utils.GenerateID(),
			UserId:    currentUser.UserId,
			RefId:     blogId,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			CreatedBy: currentUser.FullName,
			UpdatedBy: currentUser.FullName,
		}
		if err := tx.Create(&newLike).Error; err != nil {
			return err
		}
		// Increment likesCount on the associated blog
		if err := tx.Model(&models.Blog{}).Where("blog_id = ?", blogId).Update("likes_count", gorm.Expr("likes_count + ?", 1)).First(&blog).Error; err != nil {
			return err
		}

		return tx.Model(&models.UsersStats{}).Where("user_id = ?", currentUser.UserId).Update("total_likes", gorm.Expr("total_likes + 1")).Error
	})
	if err != nil {
		s.logger.Printf("Error liking/unliking blog: %v", err)
		return LikeResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to like/unlike blog"}
	}

	return LikeResponse{
		Liked:      like.LikeId == "",
		LikesCount: blog.LikesCount,
	}, nil
}

// AddComment adds a comment to a blog
func (s *BlogsService) AddComment(blogId string, dto CreateCommentDto, currentUser models.ICurrentUser) (MutationResponse, error) {

	var comment models.Comment
	err := s.db.Transaction(func(tx *gorm.DB) error {
		comment = models.Comment{
			CommentId: utils.GenerateID(),
			UserId:    currentUser.UserId,
			RefId:     blogId,
			Text:      dto.Text,
			Image:     dto.Image,
			Video:     dto.Video,
			Audio:     dto.Audio,
			Sticker:   dto.Sticker,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			CreatedBy: currentUser.FullName,
			UpdatedBy: currentUser.FullName,
		}
		if err := tx.FirstOrCreate(&comment).Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("profile_image", "full_name", "user_id", "email", "verified")
		}).Error; err != nil {
			return err
		}
		// Increment commentsCount on the associated blog
		if err := tx.Model(&models.Blog{}).Where("blog_id = ?", blogId).Update("comments_count", gorm.Expr("comments_count + ?", 1)).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.logger.Printf("Error adding comment: %v", err)
		return MutationResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to add comment"}
	}

	var user models.User
	err = s.db.Select("profile_image", "user_id", "full_name", "email", "verified").
		Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "user_id", Value: currentUser.UserId}}}).
		Preload("UserRoles.Role").First(&user).Error
	if err != nil {
		s.logger.Printf("Error fetching user details: %v", err)
		return MutationResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to fetch user details"}
	}

	var repliesCount, likesCount int64
	s.db.Model(&models.Comment{}).Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "ref_id", Value: comment.CommentId}}}).
		Count(&repliesCount)
	s.db.Model(&models.Like{}).Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "ref_id", Value: comment.CommentId}}}).
		Count(&likesCount)

	isLiked := s.db.Model(&models.Like{}).Clauses(clause.Where{
		Exprs: []clause.Expression{
			clause.Eq{Column: "ref_id", Value: comment.CommentId},
			clause.Eq{Column: "user_id", Value: currentUser.UserId},
		},
	}).RowsAffected > 0

	data := map[string]interface{}{
		"comment":      comment,
		"repliesCount": repliesCount,
		"likesCount":   likesCount,
		"liked":        isLiked,
	}

	return MutationResponse{
		Message: fmt.Sprintf("Comment added successfully to blogId = %s", blogId),
		Data:    data,
	}, nil
}

// DeleteComment deletes a comment
func (s *BlogsService) DeleteComment(commentId string, currentUser models.ICurrentUser) (map[string]string, error) {

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var comment models.Comment
		err := tx.Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "comment_id", Value: commentId}}}).
			First(&comment).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &fiber.Error{Code: fiber.StatusNotFound, Message: fmt.Sprintf("Comment with ID %s does not exist", commentId)}
			}
			return err
		}

		if err := tx.Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "ref_id", Value: commentId}}}).
			Delete(&models.Comment{}).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "ref_id", Value: commentId}}}).
			Delete(&models.Like{}).Error; err != nil {
			return err
		}
		// Decrement commentsCount on the associated blog
		if comment.RefId != "" {
			var parentBlog models.Blog
			if tx.Where("blog_id = ?", comment.RefId).First(&parentBlog).Error == nil {
				tx.Model(&parentBlog).Update("comments_count", gorm.Expr("comments_count - ?", 1))
			} else {
				// If it's a reply, decrement repliesCount on the parent comment
				tx.Model(&models.Comment{}).Where("comment_id = ?", comment.RefId).Update("replies_count", gorm.Expr("replies_count - ?", 1))
			}
		}
		return tx.Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "comment_id", Value: commentId}}}).
			Delete(&comment).Error
	})
	if err != nil {
		s.logger.Printf("Error deleting comment: %v", err)
		if fiberErr, ok := err.(*fiber.Error); ok {
			return nil, fiberErr
		}
		return nil, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to delete comment"}
	}

	return map[string]string{"message": "Comment deleted successfully"}, nil
}

// UpdateComment updates a comment
func (s *BlogsService) UpdateComment(commentId string, dto CreateCommentDto, currentUser models.ICurrentUser) (MutationResponse, error) {
	var comment models.Comment
	err := s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "comment_id", Value: commentId}}}).
			First(&comment).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &fiber.Error{Code: fiber.StatusNotFound, Message: fmt.Sprintf("Comment with ID %s does not exist", commentId)}
			}
			return err
		}

		updateData := map[string]interface{}{
			"updated_at": time.Now().UTC(),
			"updated_by": currentUser.FullName,
		}
		if dto.Text != "" {
			updateData["text"] = dto.Text
		}
		if dto.Image != "" {
			updateData["image"] = dto.Image
		}
		if dto.Video != "" {
			updateData["video"] = dto.Video
		}
		if dto.Audio != "" {
			updateData["audio"] = dto.Audio
		}
		if dto.Sticker != "" {
			updateData["sticker"] = dto.Sticker
		}

		return tx.Model(&comment).Updates(updateData).Error
	})
	if err != nil {
		s.logger.Printf("Error updating comment: %v", err)
		if fiberErr, ok := err.(*fiber.Error); ok {
			return MutationResponse{}, fiberErr
		}
		return MutationResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to update comment"}
	}

	return MutationResponse{
		Message: "Comment updated successfully",
		Data:    comment,
	}, nil
}

// func (s *BlogsService) UpdateComment(commentId string, dto CreateCommentDto, currentUser models.ICurrentUser) (MutationResponse, error) {
// 	var comment models.Comment
// 	err := s.db.Transaction(func(tx *gorm.DB) error {
// 		// First, check if comment exists and preload User
// 		err := tx.Preload("User", func(db *gorm.DB) *gorm.DB {
// 			return db.Select("profile_image", "full_name", "user_id", "email", "verified")
// 		}).Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "comment_id", Value: commentId}}}).
// 			First(&comment).Error
// 		if err != nil {
// 			if errors.Is(err, gorm.ErrRecordNotFound) {
// 				return &fiber.Error{Code: fiber.StatusNotFound, Message: fmt.Sprintf("Comment with ID %s does not exist", commentId)}
// 			}
// 			return err
// 		}

// 		// Prepare update data
// 		updateData := models.Comment{
// 			Text:      dto.Text,
// 			Image:     dto.Image,
// 			Video:     dto.Video,
// 			UpdatedAt: time.Now().UTC(),
// 			UpdatedBy: currentUser.FullName,
// 		}

// 		// Perform update
// 		err = tx.Model(&comment).Where(models.Comment{CommentId: commentId}).Updates(updateData).Error
// 		if err != nil {
// 			return err
// 		}

// 		// Fetch the updated comment
// 		return tx.Preload("User", func(db *gorm.DB) *gorm.DB {
// 			return db.Select("profile_image", "full_name", "user_id", "email", "verified")
// 		}).Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "comment_id", Value: commentId}}}).
// 			First(&comment).Error
// 	})

// 	if err != nil {
// 		s.logger.Printf("Error updating comment: %v", err)
// 		if fiberErr, ok := err.(*fiber.Error); ok {
// 			return MutationResponse{}, fiberErr
// 		}
// 		return MutationResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to update comment"}
// 	}

// 	return MutationResponse{
// 		Message: "Comment updated successfully",
// 		Data:    comment,
// 	}, nil
// }

// LikeComment handles liking or unliking a comment
func (s *BlogsService) LikeComment(commentId string, currentUser models.ICurrentUser) (LikeResponse, error) {
	if !currentUser.IsAuthenticated {
		return LikeResponse{}, &fiber.Error{Code: fiber.StatusUnauthorized, Message: "Unauthorized"}
	}

	var like models.Like
	var comment models.Comment
	err := s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Where{
			Exprs: []clause.Expression{
				clause.Eq{Column: "user_id", Value: currentUser.UserId},
				clause.Eq{Column: "ref_id", Value: commentId},
			},
		}).First(&like).Error
		if err == nil {
			// Unlike
			if err := tx.Delete(&like).Error; err != nil {
				return err
			}
			// Decrement likesCount on the associated comment
			return tx.Model(&models.Comment{}).Where("comment_id = ?", commentId).Update("likes_count", gorm.Expr("likes_count - ?", 1)).First(&comment).Error
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// Like
		newLike := models.Like{
			LikeId:    utils.GenerateID(),
			UserId:    currentUser.UserId,
			RefId:     commentId,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			CreatedBy: currentUser.FullName,
			UpdatedBy: currentUser.FullName,
		}
		if err := tx.Create(&newLike).Error; err != nil {
			return err
		}
		// Increment likesCount on the associated comment
		return tx.Model(&models.Comment{}).Where("comment_id = ?", commentId).Update("likes_count", gorm.Expr("likes_count + ?", 1)).First(&comment).Error
	})
	if err != nil {
		s.logger.Printf("Error liking/unliking comment: %v", err)
		return LikeResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to like/unlike comment"}
	}

	return LikeResponse{
		Liked:      like.LikeId == "",
		LikesCount: comment.LikesCount,
	}, nil
}

// FindCommentLikes retrieves likes for a comment
func (s *BlogsService) FindCommentLikes(commentId string, currentUser models.ICurrentUser) ([]models.Like, error) {
	var likes []models.Like
	err := s.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("profile_image", "user_id", "full_name", "email", "verified").Preload("UserRoles.Role")
	}).Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "ref_id", Value: commentId}}}).
		Find(&likes).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []models.Like{}, &fiber.Error{Code: fiber.StatusNotFound, Message: fmt.Sprintf("Comment with ID %s does not exist", commentId)}
		}
		s.logger.Printf("Error fetching comment likes: %v", err)
		return []models.Like{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to fetch likes"}
	}

	if len(likes) == 0 {
		return []models.Like{}, nil
	}

	return likes, nil
}

// FindComments retrieves comments for a blog
func (s *BlogsService) FindComments(blogId string, currentUser models.ICurrentUser, page, limit int) (models.PaginatedResponse, error) {
	var totalItems int64
	s.db.Model(&models.Comment{}).Where("ref_id = ?", blogId).Count(&totalItems)

	offset := (page - 1) * limit
	var comments []models.Comment
	err := s.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("profile_image", "user_id", "full_name", "email", "verified")
	}).Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "ref_id", Value: blogId}}}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).
		Limit(limit).Offset(offset).Find(&comments).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.PaginatedResponse{Data: []CommentWithMeta{}}, &fiber.Error{Code: fiber.StatusNotFound, Message: fmt.Sprintf("Blog with ID %s does not exist", blogId)}
		}
		s.logger.Printf("Error fetching comments: %v", err)
		return models.PaginatedResponse{Data: []CommentWithMeta{}}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to fetch comments"}
	}

	var enrichedComments []CommentWithMeta
	for _, comment := range comments {
		var liked bool = false
		if currentUser.IsAuthenticated {
			liked = s.db.Model(&models.Like{}).Clauses(clause.Where{
				Exprs: []clause.Expression{
					clause.Eq{Column: "ref_id", Value: comment.CommentId},
					clause.Eq{Column: "user_id", Value: currentUser.UserId},
				},
			}).Limit(1).Find(&models.Like{}).RowsAffected > 0
		}

		enrichedComments = append(enrichedComments, CommentWithMeta{
			Comment:      comment,
			RepliesCount: comment.RepliesCount,
			LikesCount:   comment.LikesCount,
			Liked:        liked,
		})
	}

	totalPages := (totalItems + int64(limit) - 1) / int64(limit)

	if len(enrichedComments) == 0 {
		return models.PaginatedResponse{
			Data: []CommentWithMeta{},
			Metadata: models.PaginationMetadata{
				CurrentPage:     int64(page),
				ItemsPerPage:    int64(limit),
				TotalItems:      totalItems,
				TotalPages:      totalPages,
				HasNextPage:     int64(page) < totalPages,
				HasPreviousPage: int64(page) > 1,
			},
		}, nil
	}

	return models.PaginatedResponse{
		Data: enrichedComments,
		Metadata: models.PaginationMetadata{
			CurrentPage:     int64(page),
			ItemsPerPage:    int64(limit),
			TotalItems:      totalItems,
			TotalPages:      totalPages,
			HasNextPage:     int64(page) < totalPages,
			HasPreviousPage: int64(page) > 1,
		},
	}, nil
}

// FindReplies retrieves replies for a comment
func (s *BlogsService) FindReplies(commentId string, currentUser models.ICurrentUser, page, limit int) (models.PaginatedResponse, error) {
	var totalItems int64
	s.db.Model(&models.Comment{}).Where("ref_id = ?", commentId).Count(&totalItems)

	offset := (page - 1) * limit
	var comments []models.Comment
	err := s.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("profile_image", "user_id", "full_name", "email", "verified")
	}).Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "ref_id", Value: commentId}}}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).
		Limit(limit).Offset(offset).Find(&comments).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.PaginatedResponse{Data: []CommentWithMeta{}}, &fiber.Error{Code: fiber.StatusNotFound, Message: fmt.Sprintf("Comment with ID %s does not exist", commentId)}
		}
		s.logger.Printf("Error fetching replies: %v", err)
		return models.PaginatedResponse{Data: []CommentWithMeta{}}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to fetch replies"}
	}

	var enrichedComments []CommentWithMeta
	for _, comment := range comments {
		var repliesCount, likesCount int64
		repliesCount = comment.RepliesCount
		likesCount = comment.LikesCount

		var liked bool = false

		comment.RepliesCount = repliesCount
		comment.LikesCount = likesCount

		if currentUser.IsAuthenticated {
			liked = s.db.Model(&models.Like{}).Clauses(clause.Where{
				Exprs: []clause.Expression{
					clause.Eq{Column: "ref_id", Value: comment.CommentId},
					clause.Eq{Column: "user_id", Value: currentUser.UserId},
				},
			}).Limit(1).Find(&models.Like{}).RowsAffected > 0
		}

		enrichedComments = append(enrichedComments, CommentWithMeta{
			Comment:      comment,
			RepliesCount: repliesCount,
			LikesCount:   likesCount,
			Liked:        liked,
		})
	}

	totalPages := (totalItems + int64(limit) - 1) / int64(limit)

	if len(enrichedComments) == 0 {
		return models.PaginatedResponse{
			Data: []CommentWithMeta{},
			Metadata: models.PaginationMetadata{
				CurrentPage:     int64(page),
				ItemsPerPage:    int64(limit),
				TotalItems:      totalItems,
				TotalPages:      totalPages,
				HasNextPage:     int64(page) < totalPages,
				HasPreviousPage: int64(page) > 1,
			},
		}, nil
	}

	return models.PaginatedResponse{
		Data: enrichedComments,
		Metadata: models.PaginationMetadata{
			CurrentPage:     int64(page),
			ItemsPerPage:    int64(limit),
			TotalItems:      totalItems,
			TotalPages:      totalPages,
			HasNextPage:     int64(page) < totalPages,
			HasPreviousPage: int64(page) > 1,
		},
	}, nil
}

// AddReply adds a reply to a comment
func (s *BlogsService) AddReply(commentId string, dto CreateReplyDto, currentUser models.ICurrentUser) (MutationResponse, error) {
	var reply models.Comment
	err := s.db.Transaction(func(tx *gorm.DB) error {
		reply = models.Comment{
			CommentId: utils.GenerateID(),
			RefId:     commentId,
			UserId:    currentUser.UserId,
			Text:      dto.Text,
			Image:     dto.Image,
			Video:     dto.Video,
			Sticker:   dto.Sticker,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			CreatedBy: currentUser.FullName,
			UpdatedBy: currentUser.FullName,
		}
		if err := tx.FirstOrCreate(&reply).Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("profile_image", "full_name", "user_id", "email", "verified")
		}).Error; err != nil {
			return err
		}
		// Increment repliesCount on the associated comment
		return tx.Model(&models.Comment{}).Where("comment_id = ?", commentId).Update("replies_count", gorm.Expr("replies_count + ?", 1)).Error
	})
	if err != nil {
		s.logger.Printf("Error adding reply: %v", err)
		return MutationResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to add reply"}
	}

	return MutationResponse{
		Message: fmt.Sprintf("Reply added successfully to commentId = %s", commentId),
		Data:    reply,
	}, nil
}

// FindLikesAndFollowers retrieves likes and followers for a blog
func (s *BlogsService) FindLikesAndFollowers(blogId string, currentUser models.ICurrentUser) (map[string]interface{}, error) {
	var userIds []string
	var likes []models.Like
	err := s.db.Clauses(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "ref_id", Value: blogId}}}).
		Find(&likes).Pluck("user_id", &userIds).Error
	if err != nil {
		s.logger.Printf("Error fetching likes: %v", err)
		return nil, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to fetch likes"}
	}

	var userFollowings []models.Follow
	err = s.db.Clauses(clause.Where{
		Exprs: []clause.Expression{
			clause.Or(
				clause.Eq{Column: "follower_id", Value: currentUser.UserId},
				clause.Eq{Column: "following_id", Value: currentUser.UserId},
			),
		},
	}).Find(&userFollowings).Error
	if err != nil {
		s.logger.Printf("Error fetching user followings: %v", err)
		return nil, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to fetch followings"}
	}

	var commonUserIds []interface{}
	for _, follow := range userFollowings {
		if contains(userIds, follow.FollowerId) || contains(userIds, follow.FollowingId) {
			commonUserIds = append(commonUserIds, follow.FollowerId)
		}
	}

	var users []models.User
	err = s.db.Clauses(clause.Where{Exprs: []clause.Expression{clause.IN{Column: "user_id", Values: commonUserIds}}}).
		Preload("UserRoles.Role").Find(&users).Error
	if err != nil {
		s.logger.Printf("Error fetching users: %v", err)
		return nil, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to fetch users"}
	}

	return map[string]interface{}{
		"sessionUsers": users,
	}, nil
}

// enrichBlogs enriches blogs with metadata
func (s *BlogsService) enrichBlogs(blogs []models.Blog, currentUser models.ICurrentUser) ([]BlogWithMeta, error) {
	var blogsWithMeta []BlogWithMeta = []BlogWithMeta{}
	for _, blog := range blogs {
		// Counts are now stored directly in the model and updated on creation/deletion
		var likesCount = blog.LikesCount
		var sharesCount = blog.SharesCount
		var commentsCount = blog.CommentsCount
		var viewsCount = blog.ViewsCount

		var liked bool = false
		var reposted bool = false
		if currentUser.IsAuthenticated {
			liked = s.db.Model(&models.Like{}).Clauses(clause.Where{
				Exprs: []clause.Expression{
					clause.Eq{Column: "ref_id", Value: blog.BlogId},
					clause.Eq{Column: "user_id", Value: currentUser.UserId},
				},
			}).Limit(1).Find(&models.Like{}).RowsAffected > 0
			reposted = s.db.Model(&models.Share{}).Clauses(clause.Where{
				Exprs: []clause.Expression{
					clause.Eq{Column: "ref_id", Value: blog.BlogId},
					clause.Eq{Column: "user_id", Value: currentUser.UserId},
				},
			}).Limit(1).Find(&models.Share{}).RowsAffected > 0
		}

		blogsWithMeta = append(blogsWithMeta, BlogWithMeta{
			Blog:          blog,
			Liked:         liked,
			Reposted:      reposted,
			LikesCount:    likesCount,
			RepostsCount:  sharesCount,
			CommentsCount: commentsCount,
			ViewsCount:    viewsCount,
		})
	}

	return blogsWithMeta, nil
}

// FindPinnedBlogs retrieves all currently active pinned blogs with pagination
func (s *BlogsService) FindPinnedBlogs(currentUser models.ICurrentUser, page, limit int) (models.PaginatedResponse, error) {
	now := time.Now().UTC()
	offset := (page - 1) * limit

	// Get total count of active pinned blogs
	var totalItems int64
	if err := s.db.Model(&models.PinnedBlog{}).
		Where("end_date >= ?", now).
		Count(&totalItems).Error; err != nil {
		s.logger.Printf("Error counting active pinned blogs: %v", err)
		return models.PaginatedResponse{}, &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unable to count pinned blogs",
		}
	}

	// Fetch active pinned blogs with pagination
	var pinnedBlogs []models.PinnedBlog
	err := s.db.
		Where("end_date >= ?", now).
		Preload("Blog.User", func(db *gorm.DB) *gorm.DB {
			return db.Select("profile_image", "full_name", "user_id", "email", "verified")
		}).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&pinnedBlogs).Error

	if err != nil {
		s.logger.Printf("Error fetching pinned blogs: %v", err)
		return models.PaginatedResponse{}, &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unable to fetch pinned blogs",
		}
	}

	// Extract blogs and filter out nil entries
	blogs := make([]models.Blog, 0, len(pinnedBlogs))
	for _, pb := range pinnedBlogs {
		if pb.Blog != nil {
			blogs = append(blogs, *pb.Blog)
		}
	}

	// Enrich blogs with metadata
	blogsWithMeta, err := s.enrichBlogs(blogs, currentUser)
	if err != nil {
		return models.PaginatedResponse{}, fmt.Errorf("failed to enrich blogs: %w", err)
	}

	// Calculate pagination metadata
	totalPages := (totalItems + int64(limit) - 1) / int64(limit)

	return models.PaginatedResponse{
		Data: blogsWithMeta,
		Metadata: models.PaginationMetadata{
			CurrentPage:     int64(page),
			ItemsPerPage:    int64(limit),
			TotalItems:      totalItems,
			TotalPages:      totalPages,
			HasNextPage:     int64(page) < totalPages,
			HasPreviousPage: int64(page) > 1,
		},
	}, nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
