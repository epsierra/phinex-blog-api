package middlewares

import (
	"strings"

	"github.com/epsierra/phinex-blog-api/src/utils" // Import utility functions
	"github.com/gofiber/fiber/v2"
)

type Locals struct {
	UserId          string            `json:"userId"`
	DeviceId        string            `json:"deviceId"`
	IsAuthenticated bool              `json:"isAuthenticated"`
	Keys            map[string]string `json:"keys"`
}

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract the authorization header
		authorization := c.Get("Authorization")
		accessToken := ""

		// If the header contains "Bearer", extract the token
		if authorization != "" {
			parts := strings.Split(authorization, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				accessToken = parts[1]
			}
		}

		// Get the current request URL
		url := c.OriginalURL()

		// Skip the authentication check for these routes
		if contains(url, []string{"metrics", "bloggers", "sse", "webhook"}) {
			return c.Next()
		}

		// If accessToken is provided, decode and verify
		if accessToken != "" {
			// Use utility function to decode the JWT token and return decoded data
			decodedData, err := utils.JwtDecode(accessToken)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "Invalid or expired token",
				})
			}

			// Set the decoded data in locals
			c.Locals("userId", decodedData["userId"])
			c.Locals("deviceId", decodedData["deviceId"])
			c.Locals("isAuthenticated", decodedData["isAuthenticated"])
			c.Locals("token", accessToken)

			// Proceed to the next middleware or handler
			return c.Next()
		}

		// If no token is found, return an unauthorized error
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Authorization token is required",
		})
	}
}

// Helper function to check if a URL matches any of the provided substrings
func contains(url string, substrings []string) bool {
	for _, substring := range substrings {
		if strings.Contains(url, substring) {
			return true
		}
	}
	return false
}
