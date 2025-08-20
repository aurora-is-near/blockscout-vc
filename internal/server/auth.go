package server

import (
	"encoding/base64"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

// Basic authentication middleware
func authMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get credentials from config
		username := viper.GetString("auth.username")
		password := viper.GetString("auth.password")

		// If no credentials are set, allow all requests (development mode)
		if username == "" || password == "" {
			return c.Next()
		}

		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		// Check if it's Basic auth
		if !strings.HasPrefix(authHeader, "Basic ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization format. Use Basic authentication",
			})
		}

		// Extract and decode credentials
		encodedCredentials := strings.TrimPrefix(authHeader, "Basic ")
		decodedCredentials, err := base64.StdEncoding.DecodeString(encodedCredentials)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		// Parse username:password
		credentials := strings.SplitN(string(decodedCredentials), ":", 2)
		if len(credentials) != 2 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials format",
			})
		}

		// Validate credentials
		if credentials[0] != username || credentials[1] != password {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Invalid username or password",
			})
		}

		// Credentials are valid, proceed to next middleware/handler
		return c.Next()
	}
}
