package users

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/epsierra/phinex-blog-api/src/models"
	"github.com/epsierra/phinex-blog-api/src/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UsersService handles user-related operations
type UsersService struct {
	db     *gorm.DB
	logger *log.Logger
}

// NewUsersService creates a new UsersService instance
func NewUsersService(db *gorm.DB) *UsersService {
	return &UsersService{
		db:     db,
		logger: log.New(os.Stderr, "users-service: ", log.LstdFlags),
	}
}

// CreateUser creates a new user
func (s *UsersService) CreateUser(dto CreateUserDto, currentUser models.ICurrentUser) (UserResponse, error) {
	var user models.User

	// Hash the user's password before saving
	hashedPassword, err := utils.HashData(dto.Password)
	if err != nil {
		s.logger.Printf("Error hashing password: %v", err)
		return UserResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to hash password"}
	}

	user = models.User{
		UserId:    utils.GenerateID(),
		FullName:  dto.FullName,
		Email:     dto.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		CreatedBy: currentUser.FullName,
		UpdatedBy: currentUser.FullName,
	}

	if err := s.db.Create(&user).Error; err != nil {
		s.logger.Printf("Error creating user: %v", err)
		return UserResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to add user"}
	}

	if err := s.createUserStats(user.UserId, currentUser); err != nil {
		return UserResponse{}, err
	}

	return UserResponse{
		Message: "User created successfully",
		Data:    user,
	}, nil
}

func (s *UsersService) createUserStats(userId string, currentUser models.ICurrentUser) error {
	userStats := models.UsersStats{
		UserStatsID: utils.GenerateID(),
		UserID:      userId,
		CreatedBy:   currentUser.FullName,
		UpdatedBy:   currentUser.FullName,
	}

	if err := s.db.Create(&userStats).Error; err != nil {
		s.logger.Printf("Error creating user stats: %v", err)
		return &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to create user stats"}
	}

	return nil
}

// FindAllUsers retrieves all users with pagination and optional search, ordered randomly with seed
func (s *UsersService) FindAllUsers(page, limit int, search string, currentUser models.ICurrentUser) (models.PaginatedResponse, error) {
	// Set default values for page and limit if not provided or invalid
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	// Set random seed based on current time for reproducible results
	rand.Seed(time.Now().UnixNano())

	var totalItems int64
	query := s.db.Model(&models.User{})
	if search != "" && len(search) <= 100 && search != "undefined" {
		// Normalize search term: trim and split into words
		searchTerms := strings.Fields(strings.TrimSpace(search))
		for _, term := range searchTerms {
			if term != "" {
				searchPattern := "%" + term + "%"
				query = query.Where("full_name LIKE ? OR email LIKE ? OR user_name LIKE ? OR bio LIKE ?",
					searchPattern, searchPattern, searchPattern, searchPattern)
			}
		}
	}
	query.Count(&totalItems)

	offset := (page - 1) * limit
	var users []models.User
	if err := query.Order("RANDOM()").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		s.logger.Printf("Error fetching users: %v", err)
		return models.PaginatedResponse{Data: []models.User{}}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to fetch users"}
	}

	enrichedUsers, err := s.enrichUsersWithFollowing(users, currentUser)
	if err != nil {
		return models.PaginatedResponse{Data: []models.User{}}, err
	}

	totalPages := (totalItems + int64(limit) - 1) / int64(limit)

	return models.PaginatedResponse{
		Data: enrichedUsers,
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

// FindUserById retrieves a single user by ID
func (s *UsersService) FindUserById(userId string, currentUser models.ICurrentUser) (models.User, error) {
	var user models.User
	if err := s.db.Preload("UsersStats").Where(&models.User{UserId: userId}).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, &fiber.Error{Code: fiber.StatusNotFound, Message: fmt.Sprintf("User with ID %s not found", userId)}
		}
		s.logger.Printf("Error fetching user: %v", err)
		return models.User{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to fetch user"}
	}

	if currentUser.IsAuthenticated {
		var follow models.Follow
		err := s.db.Where(&models.Follow{FollowerId: currentUser.UserId, FollowingId: user.UserId}).First(&follow).Error
		if err == nil {
			user.Following = true
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, err
		}
	}

	return user, nil
}

// UpdateUser updates an existing user
func (s *UsersService) UpdateUser(userId string, dto UpdateUserDto, currentUser models.ICurrentUser) (UserResponse, error) {
	var user models.User
	if err := s.db.Where(&models.User{UserId: userId}).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return UserResponse{}, &fiber.Error{Code: fiber.StatusNotFound, Message: fmt.Sprintf("User with ID %s not found", userId)}
		}
		s.logger.Printf("Error fetching user for update: %v", err)
		return UserResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to update user"}
	}

	updateData := map[string]interface{}{
		"updated_at": time.Now().UTC(),
		"updated_by": currentUser.FullName,
	}

	if dto.FullName != "" {
		updateData["full_name"] = dto.FullName
	}
	if dto.Email != "" {
		updateData["email"] = dto.Email
	}
	if dto.Password != "" {
		hashedPassword, err := utils.HashData(dto.Password)
		if err != nil {
			s.logger.Printf("Error hashing password for update: %v", err)
			return UserResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to hash password"}
		}
		updateData["password"] = hashedPassword
	}

	if err := s.db.Model(&user).Updates(updateData).Error; err != nil {
		s.logger.Printf("Error updating user: %v", err)
		return UserResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to update user"}
	}

	user.Password = ""
	return UserResponse{
		Message: "User updated successfully",
		Data:    user,
	}, nil
}

// DeleteUser deletes a user
func (s *UsersService) DeleteUser(userId string, currentUser models.ICurrentUser) (UserResponse, error) {
	var user models.User
	if err := s.db.Where(&models.User{UserId: userId}).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return UserResponse{}, &fiber.Error{Code: fiber.StatusNotFound, Message: fmt.Sprintf("User with ID %s not found", userId)}
		}
		s.logger.Printf("Error fetching user for deletion: %v", err)
		return UserResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to delete user"}
	}

	if err := s.db.Delete(&user).Error; err != nil {
		s.logger.Printf("Error deleting user: %v", err)
		return UserResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to delete user"}
	}

	return UserResponse{
		Message: "User deleted successfully",
	}, nil
}

// FollowUnfollowUser allows a user to follow or unfollow another user
func (s *UsersService) FollowUnfollowUser(dto FollowUnfollowDto, currentUser models.ICurrentUser) (FollowResponse, error) {
	var follow models.Follow
	err := s.db.Where(&models.Follow{FollowerId: dto.FollowerId, FollowingId: dto.FollowingId}).First(&follow).Error

	if err == nil {
		// If a follow record exists, unfollow the user
		if err := s.db.Delete(&follow).Error; err != nil {
			s.logger.Printf("Error unfollowing user: %v", err)
			return FollowResponse{}, &fiber.Error{Code: fiber.StatusUnprocessableEntity, Message: "Failed to unfollow"}
		}

		// Decrement followers_count for the user being unfollowed
		if err := s.db.Model(&models.UsersStats{}).Where("user_id = ?", dto.FollowingId).Update("followers_count", gorm.Expr("followers_count - 1")).Error; err != nil {
			s.logger.Printf("Error updating followers count: %v", err)
		}

		// Decrement followings_count for the user who is unfollowing
		if err := s.db.Model(&models.UsersStats{}).Where("user_id = ?", dto.FollowerId).Update("followings_count", gorm.Expr("followings_count - 1")).Error; err != nil {
			s.logger.Printf("Error updating followings count: %v", err)
		}

		return FollowResponse{
			Followed: false,
		}, nil
	}

	// If no follow record exists, follow the user
	follow = models.Follow{
		FollowId:    utils.GenerateID(),
		FollowerId:  dto.FollowerId,
		FollowingId: dto.FollowingId,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		CreatedBy:   currentUser.FullName,
		UpdatedBy:   currentUser.FullName,
	}

	if err := s.db.Create(&follow).Error; err != nil {
		s.logger.Printf("Error following user: %v", err)
		return FollowResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to follow"}
	}

	// Increment followers_count for the user being followed
	if err := s.db.Model(&models.UsersStats{}).Where("user_id = ?", dto.FollowingId).Update("followers_count", gorm.Expr("followers_count + 1")).Error; err != nil {
		s.logger.Printf("Error updating followers count: %v", err)
	}

	// Increment followings_count for the user who is following
	if err := s.db.Model(&models.UsersStats{}).Where("user_id = ?", dto.FollowerId).Update("followings_count", gorm.Expr("followings_count + 1")).Error; err != nil {
		s.logger.Printf("Error updating followings count: %v", err)
	}

	return FollowResponse{
		Followed: true,
	}, nil
}

// FindUsersNotFollowing retrieves users who are not followed by a specific user, ordered randomly with seed
func (s *UsersService) FindUsersNotFollowing(userId string, page, limit int, search string, currentUser models.ICurrentUser) (models.PaginatedResponse, error) {
	// Set default values for page and limit if not provided or invalid
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	// Set random seed based on current time for reproducible results
	rand.Seed(time.Now().UnixNano())

	var totalItems int64
	query := s.db.Model(&models.User{})
	if search != "" && len(search) <= 100 && search != "undefined" {
		// Normalize search term: trim and split into words
		searchTerms := strings.Fields(strings.TrimSpace(search))
		for _, term := range searchTerms {
			if term != "" {
				searchPattern := "%" + term + "%"
				query = query.Where("full_name LIKE ? OR email LIKE ? OR user_name LIKE ? OR bio LIKE ?",
					searchPattern, searchPattern, searchPattern, searchPattern)
			}
		}
	}
	query.Count(&totalItems)

	offset := (page - 1) * limit
	var ids []interface{}
	s.db.Model(&models.Follow{}).Where(&models.Follow{FollowerId: userId}).Pluck("following_id", &ids)

	var users []models.User
	query = s.db.Clauses(clause.Where{Exprs: []clause.Expression{clause.Not(clause.IN{Column: "user_id", Values: ids})}})
	if search != "" && len(search) <= 100 && search != "undefined" {
		// Normalize search term: trim and split into words
		searchTerms := strings.Fields(strings.TrimSpace(search))
		for _, term := range searchTerms {
			if term != "" {
				searchPattern := "%" + term + "%"
				query = query.Where("full_name LIKE ? OR email LIKE ? OR user_name LIKE ? OR bio LIKE ?",
					searchPattern, searchPattern, searchPattern, searchPattern)
			}
		}
	}
	err := query.Order("RANDOM()").Limit(limit).Offset(offset).Find(&users).Error

	if err != nil {
		s.logger.Printf("Error fetching users not following: %v", err)
		return models.PaginatedResponse{Data: []models.User{}}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to fetch users"}
	}

	enrichedUsers, err := s.enrichUsersWithFollowing(users, currentUser)
	if err != nil {
		return models.PaginatedResponse{Data: []models.User{}}, err
	}

	totalPages := (totalItems + int64(limit) - 1) / int64(limit)

	return models.PaginatedResponse{
		Data: enrichedUsers,
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

// FindUserFollowers retrieves the followers of a specific user, ordered randomly with seed
func (s *UsersService) FindUserFollowers(userId string, page, limit int, search string, currentUser models.ICurrentUser) (models.PaginatedResponse, error) {
	// Set default values for page and limit if not provided or invalid
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	// Set random seed based on current time for reproducible results
	rand.Seed(time.Now().UnixNano())

	var totalItems int64
	query := s.db.Model(&models.Follow{}).Where(&models.Follow{FollowingId: userId})
	if search != "" && len(search) <= 100 && search != "undefined" {
		// Normalize search term: trim and split into words
		searchTerms := strings.Fields(strings.TrimSpace(search))
		query = query.Joins("JOIN users ON users.user_id = follows.follower_id")
		for _, term := range searchTerms {
			if term != "" {
				searchPattern := "%" + term + "%"
				query = query.Where("users.full_name LIKE ? OR users.email LIKE ? OR users.user_name LIKE ? OR users.bio LIKE ?",
					searchPattern, searchPattern, searchPattern, searchPattern)
			}
		}
	}
	query.Count(&totalItems)

	offset := (page - 1) * limit
	var followerIds []interface{}
	s.db.Model(&models.Follow{}).Where(&models.Follow{FollowingId: userId}).Pluck("follower_id", &followerIds)

	var users []models.User
	query = s.db.Clauses(clause.Where{Exprs: []clause.Expression{clause.IN{Column: "user_id", Values: followerIds}}})
	if search != "" && len(search) <= 100 && search != "undefined" {
		// Normalize search term: trim and split into words
		searchTerms := strings.Fields(strings.TrimSpace(search))
		for _, term := range searchTerms {
			if term != "" {
				searchPattern := "%" + term + "%"
				query = query.Where("full_name LIKE ? OR email LIKE ? OR user_name LIKE ? OR bio LIKE ?",
					searchPattern, searchPattern, searchPattern, searchPattern)
			}
		}
	}
	err := query.Order("RANDOM()").Limit(limit).Offset(offset).Find(&users).Error

	if err != nil {
		s.logger.Printf("Error fetching user followers: %v", err)
		return models.PaginatedResponse{Data: []models.User{}}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to fetch followers"}
	}

	enrichedUsers, err := s.enrichUsersWithFollowing(users, currentUser)
	if err != nil {
		return models.PaginatedResponse{Data: []models.User{}}, err
	}

	totalPages := (totalItems + int64(limit) - 1) / int64(limit)

	return models.PaginatedResponse{
		Data: enrichedUsers,
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

// FindUserFollowings retrieves a user's followings, ordered randomly with seed
func (s *UsersService) FindUserFollowings(userId string, page, limit int, search string, currentUser models.ICurrentUser) (models.PaginatedResponse, error) {
	// Set default values for page and limit if not provided or invalid
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	// Set random seed based on current time for reproducible results
	rand.Seed(time.Now().UnixNano())

	var totalItems int64
	query := s.db.Model(&models.Follow{}).Where(&models.Follow{FollowerId: userId})
	if search != "" && len(search) <= 100 && search != "undefined" {
		// Normalize search term: trim and split into words
		searchTerms := strings.Fields(strings.TrimSpace(search))
		query = query.Joins("JOIN users ON users.user_id = follows.following_id")
		for _, term := range searchTerms {
			if term != "" {
				searchPattern := "%" + term + "%"
				query = query.Where("users.full_name LIKE ? OR users.email LIKE ? OR users.user_name LIKE ? OR users.bio LIKE ?",
					searchPattern, searchPattern, searchPattern, searchPattern)
			}
		}
	}
	query.Count(&totalItems)

	offset := (page - 1) * limit
	var followingIds []interface{}
	s.db.Model(&models.Follow{}).Where(&models.Follow{FollowerId: userId}).Pluck("following_id", &followingIds)

	var users []models.User
	query = s.db.Clauses(clause.Where{Exprs: []clause.Expression{clause.IN{Column: "user_id", Values: followingIds}}})
	if search != "" && len(search) <= 100 && search != "undefined" {
		// Normalize search term: trim and split into words
		searchTerms := strings.Fields(strings.TrimSpace(search))
		for _, term := range searchTerms {
			if term != "" {
				searchPattern := "%" + term + "%"
				query = query.Where("full_name LIKE ? OR email LIKE ? OR user_name LIKE ? OR bio LIKE ?",
					searchPattern, searchPattern, searchPattern, searchPattern)
			}
		}
	}
	err := query.Order("RANDOM()").Limit(limit).Offset(offset).Find(&users).Error

	if err != nil {
		s.logger.Printf("Error fetching user followings: %v", err)
		return models.PaginatedResponse{Data: []models.User{}}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to fetch followings"}
	}

	enrichedUsers, err := s.enrichUsersWithFollowing(users, currentUser)
	if err != nil {
		return models.PaginatedResponse{Data: []models.User{}}, err
	}

	totalPages := (totalItems + int64(limit) - 1) / int64(limit)

	return models.PaginatedResponse{
		Data: enrichedUsers,
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

// enrichUsersWithFollowing enriches users with following status
func (s *UsersService) enrichUsersWithFollowing(users []models.User, currentUser models.ICurrentUser) ([]models.User, error) {
	if !currentUser.IsAuthenticated {
		return users, nil
	}
	var enrichedUsers []models.User = []models.User{}
	for _, user := range users {
		var follow models.Follow
		err := s.db.Model(models.Follow{}).Where(&models.Follow{FollowerId: currentUser.UserId, FollowingId: user.UserId}).First(&follow).Error
		if err == nil {
			user.Following = true
		} else {
			user.Following = false
		}
		enrichedUsers = append(enrichedUsers, user)
	}
	return enrichedUsers, nil
}
