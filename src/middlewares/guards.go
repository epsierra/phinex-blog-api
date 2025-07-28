package middlewares

import (
	"strings"

	"github.com/epsierra/phinex-blog-api/src/models" // Adjust to your project structure
	"github.com/epsierra/phinex-blog-api/src/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Helper function to check if a URL matches any of the provided substrings
func contains(url string, substrings []string) bool {
	for _, substring := range substrings {
		if strings.Contains(url, substring) {
			return true
		}
	}
	return false
}

func AdminGuard(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Extract token from Authorization header
		authorization := c.Get("Authorization")
		if authorization == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "No token provided",
			})
		}

		parts := strings.Split(authorization, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid token format",
			})
		}
		tokenStr := parts[1]

		// Decode and verify JWT
		decodedData, err := utils.JwtDecode(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid or expired token",
			})
		}

		userId, ok := decodedData["userId"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid user ID in token",
			})
		}

		// Query user with roles
		var user models.User
		err = db.Preload("UserRoles.Role").Where("user_id = ?", userId).First(&user).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "Invalid token",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error",
			})
		}

		// Check if user has Admin or SuperAdmin role
		isAdminOrHigher := false
		roles := make([]string, 0, len(user.UserRoles))
		for _, ur := range user.UserRoles {
			roleName := string(ur.Role.RoleName)
			roles = append(roles, roleName)
			if roleName == string(models.RoleNameAdmin) || roleName == string(models.RoleNameSuperAdmin) {
				isAdminOrHigher = true
			}
		}

		if !isAdminOrHigher {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Forbidden: Admins or super admins only",
			})
		}

		// Set current user in context
		currentUser := models.ICurrentUser{
			UserId:          user.UserId,
			Email:           user.Email,
			FullName:        user.FullName,
			Roles:           roles,
			IsAuthenticated: true,
			IP:              c.IP(),
		}
		c.Locals("user", currentUser)

		return c.Next()
	}
}

func AnonymousGuard(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip authentication for specific routes
		url := c.OriginalURL()
		if contains(url, []string{"metrics", "bloggers", "sse", "webhook", "swagger-docs"}) {
			return c.Next()
		}

		// Extract token from Authorization header
		authorization := c.Get("Authorization")
		if authorization == "" {
			// Allow anonymous access
			currentUser := models.ICurrentUser{
				Roles:           []string{string(models.RoleNameAnonymous)},
				IsAuthenticated: false,
				IP:              c.IP(),
			}
			c.Locals("user", currentUser)
			return c.Next()
		}

		parts := strings.Split(authorization, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid token format",
			})
		}
		tokenStr := parts[1]

		// Decode and verify JWT
		decodedData, err := utils.JwtDecode(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid or expired token",
			})
		}

		userId, ok := decodedData["userId"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid user ID in token",
			})
		}

		// Query user with roles
		var user models.User
		err = db.Preload("UserRoles.Role").Where("user_id = ?", userId).First(&user).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "Invalid token",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error",
			})
		}

		// Check if user has any valid role
		validRoles := []string{
			string(models.RoleNameAdmin),
			string(models.RoleNameSuperAdmin),
			string(models.RoleNameBusinessOwner),
			string(models.RoleNamePaymentAgent),
			string(models.RoleNameAuthenticated),
			string(models.RoleNameAnonymous),
		}
		isValidRole := false
		roles := make([]string, 0, len(user.UserRoles))
		for _, ur := range user.UserRoles {
			roleName := string(ur.Role.RoleName)
			roles = append(roles, roleName)
			for _, validRole := range validRoles {
				if roleName == validRole {
					isValidRole = true
					break
				}
			}
		}

		if !isValidRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Forbidden: No valid role assigned",
			})
		}

		// Set current user in context
		currentUser := models.ICurrentUser{
			UserId:          user.UserId,
			Email:           user.Email,
			FullName:        user.FullName,
			Roles:           roles,
			IsAuthenticated: false,
			IP:              c.IP(),
		}
		c.Locals("user", currentUser)

		return c.Next()
	}
}

func AuthenticatedGuard(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip authentication for specific routes
		// url := c.OriginalURL()
		// if contains(url, []string{"metrics", "bloggers", "sse", "webhook", "swagger-docs"}) {
		// 	return c.Next()
		// }

		// Extract token from Authorization header
		authorization := c.Get("Authorization")
		if authorization == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "No token provided",
			})
		}

		parts := strings.Split(authorization, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid token format",
			})
		}
		tokenStr := parts[1]

		// Decode and verify JWT
		decodedData, err := utils.JwtDecode(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid or expired token",
			})
		}

		userId, ok := decodedData["userId"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid user ID in token",
			})
		}

		// Query user with roles
		var user models.User
		err = db.Preload("UserRoles.Role").Where("user_id = ?", userId).First(&user).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "Invalid token",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error",
			})
		}

		// Check if user is authenticated
		validRoles := []string{
			string(models.RoleNameAdmin),
			string(models.RoleNameSuperAdmin),
			string(models.RoleNameBusinessOwner),
			string(models.RoleNamePaymentAgent),
			string(models.RoleNameAuthenticated),
		}
		isAuthenticated := false
		roles := make([]string, 0, len(user.UserRoles))
		for _, ur := range user.UserRoles {
			roleName := string(ur.Role.RoleName)
			roles = append(roles, roleName)
			for _, validRole := range validRoles {
				if roleName == validRole {
					isAuthenticated = true
					break
				}
			}
		}

		// if !isAuthenticated {
		// 	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
		// 		"message": "Forbidden: Authenticated users only",
		// 	})
		// }

		// Set current user in context
		currentUser := models.ICurrentUser{
			UserId:          user.UserId,
			Email:           user.Email,
			FullName:        user.FullName,
			Roles:           roles,
			IsAuthenticated: isAuthenticated,
			IP:              c.IP(),
		}
		c.Locals("user", currentUser)

		return c.Next()
	}
}

func BusinessOwnerGuard(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip authentication for specific routes
		// url := c.OriginalURL()
		// if contains(url, []string{"metrics", "bloggers", "sse", "webhook"}) {
		// 	return c.Next()
		// }

		// Extract token from Authorization header
		authorization := c.Get("Authorization")
		if authorization == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "No token provided",
			})
		}

		parts := strings.Split(authorization, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid token format",
			})
		}
		tokenStr := parts[1]

		// Decode and verify JWT
		decodedData, err := utils.JwtDecode(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid or expired token",
			})
		}

		userId, ok := decodedData["userId"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid user ID in token",
			})
		}

		// Query user with roles
		var user models.User
		err = db.Preload("UserRoles.Role").Where("user_id = ?", userId).First(&user).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "Invalid token",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error",
			})
		}

		// Check if user has BusinessOwner, Admin, or SuperAdmin role
		isBusinessOwnerOrHigher := false
		roles := make([]string, 0, len(user.UserRoles))
		for _, ur := range user.UserRoles {
			roleName := string(ur.Role.RoleName)
			roles = append(roles, roleName)
			if roleName == string(models.RoleNameBusinessOwner) ||
				roleName == string(models.RoleNameAdmin) ||
				roleName == string(models.RoleNameSuperAdmin) {
				isBusinessOwnerOrHigher = true
			}
		}

		if !isBusinessOwnerOrHigher {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Forbidden: Business owners, admins, or super admins only",
			})
		}

		// Set current user in context
		currentUser := models.ICurrentUser{
			UserId:          user.UserId,
			Email:           user.Email,
			FullName:        user.FullName,
			Roles:           roles,
			IsAuthenticated: true,
			IP:              c.IP(),
		}
		c.Locals("user", currentUser)

		return c.Next()
	}
}

func PaymentAgentGuard(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip authentication for specific routes
		// url := c.OriginalURL()
		// if contains(url, []string{"metrics", "bloggers", "sse", "webhook"}) {
		// 	return c.Next()
		// }

		// Extract token from Authorization header
		authorization := c.Get("Authorization")
		if authorization == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "No token provided",
			})
		}

		parts := strings.Split(authorization, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid token format",
			})
		}
		tokenStr := parts[1]

		// Decode and verify JWT
		decodedData, err := utils.JwtDecode(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid or expired token",
			})
		}

		userId, ok := decodedData["userId"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid user ID in token",
			})
		}

		// Query user with roles
		var user models.User
		err = db.Preload("UserRoles.Role").Where("user_id = ?", userId).First(&user).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "Invalid token",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error",
			})
		}

		// Check if user has PaymentAgent, Admin, or SuperAdmin role
		isPaymentAgentOrHigher := false
		roles := make([]string, 0, len(user.UserRoles))
		for _, ur := range user.UserRoles {
			roleName := string(ur.Role.RoleName)
			roles = append(roles, roleName)
			if roleName == string(models.RoleNamePaymentAgent) ||
				roleName == string(models.RoleNameAdmin) ||
				roleName == string(models.RoleNameSuperAdmin) {
				isPaymentAgentOrHigher = true
			}
		}

		if !isPaymentAgentOrHigher {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Forbidden: Payment agents, admins, or super admins only",
			})
		}

		// Set current user in context
		currentUser := models.ICurrentUser{
			UserId:          user.UserId,
			Email:           user.Email,
			FullName:        user.FullName,
			Roles:           roles,
			IsAuthenticated: true,
			IP:              c.IP(),
		}
		c.Locals("user", currentUser)

		return c.Next()
	}
}

func SuperAdminGuard(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip authentication for specific routes
		// url := c.OriginalURL()
		// if contains(url, []string{"metrics", "bloggers", "sse", "webhook"}) {
		// 	return c.Next()
		// }

		// Extract token from Authorization header
		authorization := c.Get("Authorization")
		if authorization == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "No token provided",
			})
		}

		parts := strings.Split(authorization, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid token format",
			})
		}
		tokenStr := parts[1]

		// Decode and verify JWT
		decodedData, err := utils.JwtDecode(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid or expired token",
			})
		}

		userId, ok := decodedData["userId"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid user ID in token",
			})
		}

		// Query user with roles
		var user models.User
		err = db.Preload("UserRoles.Role").Where("user_id = ?", userId).First(&user).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "Invalid token",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error",
			})
		}

		// Check if user has SuperAdmin role
		isSuperAdmin := false
		roles := make([]string, 0, len(user.UserRoles))
		for _, ur := range user.UserRoles {
			roleName := string(ur.Role.RoleName)
			roles = append(roles, roleName)
			if roleName == string(models.RoleNameSuperAdmin) {
				isSuperAdmin = true
			}
		}

		if !isSuperAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Forbidden: Super admins only",
			})
		}

		// Set current user in context
		currentUser := models.ICurrentUser{
			UserId:          user.UserId,
			Email:           user.Email,
			FullName:        user.FullName,
			Roles:           roles,
			IsAuthenticated: true,
			IP:              c.IP(),
		}
		c.Locals("user", currentUser)

		return c.Next()
	}
}
