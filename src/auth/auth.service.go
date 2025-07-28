package auth

import (
	"errors"
	"log"
	"os"

	"github.com/epsierra/phinex-blog-api/src/models"
	"github.com/epsierra/phinex-blog-api/src/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// AuthService handles authentication-related operations
type AuthService struct {
	db        *gorm.DB
	logger    *log.Logger
	jwtSecret string
}

// NewAuthService creates a new AuthService instance
func NewAuthService(db *gorm.DB) *AuthService {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable not set")
	}
	return &AuthService{
		db:        db,
		logger:    log.New(os.Stderr, "auth-service: ", log.LstdFlags),
		jwtSecret: jwtSecret,
	}
}

// GetTokenByEmail generates JWT token for a verified user by email.
func (s *AuthService) GetTokenByEmail(email string) (TokenResponse, error) {
	var user models.User
	err := s.db.Preload("UserRoles.Role").Where(&models.User{Email: email}).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TokenResponse{}, &fiber.Error{Code: fiber.StatusNotFound, Message: "User not found"}
		}
		s.logger.Printf("Error fetching user by email: %v", err)
		return TokenResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to get token"}
	}

	if !user.Verified {
		return TokenResponse{}, &fiber.Error{Code: fiber.StatusUnauthorized, Message: "User is not verified"}
	}

	// Prepare claims
	roles := []string{}
	for _, userRole := range user.UserRoles {
		roles = append(roles, string(userRole.Role.RoleName))
	}

	tokenString, err := utils.JwtEncode(map[string]interface{}{
		"userId":          user.UserId,
		"isAuthenticated": true,
		"email":           user.Email,
		"roles":           roles,
	})
	if err != nil {
		s.logger.Printf("Error signing token: %v", err)
		return TokenResponse{}, &fiber.Error{Code: fiber.StatusInternalServerError, Message: "Failed to generate token"}
	}

	// Exclude password from user object
	user.Password = ""

	return TokenResponse{
		Message: "Token generated successfully",
		Token:   tokenString,
		User:    user,
	}, nil
}
