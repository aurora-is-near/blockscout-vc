package server

import (
	"crypto/subtle"
	"encoding/base64"
	"strings"

	"blockscout-vc/internal/config"

	"github.com/gofiber/fiber/v2"
)

// Basic authentication middleware
func authMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get credentials from config using project's config getters
		username := config.GetAuthUsername()
		password := config.GetAuthPassword()

		// Only disable auth if BOTH username AND password are empty
		// This prevents fail-open on partial/misconfigured secrets
		if username == "" && password == "" {
			return c.Next()
		}

		// If only one credential is set, require both to be present
		if username == "" || password == "" {
			c.Status(fiber.StatusUnauthorized)
			c.Set("WWW-Authenticate", `Basic realm="Restricted"`)
			return c.JSON(fiber.Map{
				"error": "Authentication required - both username and password must be configured",
			})
		}

		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			c.Status(fiber.StatusUnauthorized)
			c.Set("WWW-Authenticate", `Basic realm="Restricted"`)
			return c.JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		// Parse Authorization header to extract scheme and credentials
		// Use case-insensitive scheme comparison as per RFC 7235
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 {
			c.Status(fiber.StatusUnauthorized)
			c.Set("WWW-Authenticate", `Basic realm="Restricted"`)
			return c.JSON(fiber.Map{
				"error": "Invalid authorization format. Use Basic authentication",
			})
		}

		scheme := parts[0]
		encodedCredentials := parts[1]

		// Case-insensitive scheme check
		if !strings.EqualFold(scheme, "Basic") {
			c.Status(fiber.StatusUnauthorized)
			c.Set("WWW-Authenticate", `Basic realm="Restricted"`)
			return c.JSON(fiber.Map{
				"error": "Invalid authorization scheme. Use Basic authentication",
			})
		}

		// Extract and decode credentials
		decodedCredentials, err := base64.StdEncoding.DecodeString(encodedCredentials)
		if err != nil {
			c.Status(fiber.StatusUnauthorized)
			c.Set("WWW-Authenticate", `Basic realm="Restricted"`)
			return c.JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		// Parse username:password
		credentials := strings.SplitN(string(decodedCredentials), ":", 2)
		if len(credentials) != 2 {
			c.Status(fiber.StatusUnauthorized)
			c.Set("WWW-Authenticate", `Basic realm="Restricted"`)
			return c.JSON(fiber.Map{
				"error": "Invalid credentials format",
			})
		}

		// Validate credentials using constant-time comparison to prevent timing attacks
		usernameMatch := subtle.ConstantTimeCompare([]byte(credentials[0]), []byte(username)) == 1
		passwordMatch := subtle.ConstantTimeCompare([]byte(credentials[1]), []byte(password)) == 1

		if !usernameMatch || !passwordMatch {
			c.Status(fiber.StatusUnauthorized)
			c.Set("WWW-Authenticate", `Basic realm="Restricted"`)
			return c.JSON(fiber.Map{
				"error": "Invalid username or password",
			})
		}

		// Credentials are valid, proceed to next middleware/handler
		return c.Next()
	}
}
