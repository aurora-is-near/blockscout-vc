package server

import (
	"blockscout-vc/internal/client"
	"blockscout-vc/internal/database"
	"blockscout-vc/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"blockscout-vc/internal/config"
)

// getStringValue safely extracts string value from sql.NullString
func getStringValue(nullString sql.NullString) string {
	if nullString.Valid {
		return nullString.String
	}
	return ""
}

type Server struct {
	app              *fiber.App
	database         *database.Database
	blockscoutClient *client.BlockscoutClient
}

func NewServer() (*Server, error) {
	app := fiber.New(fiber.Config{
		AppName: "Blockscout VC API",
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: strings.Join(config.GetCORSAllowedOrigins(), ","),
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Initialize database
	db, err := database.NewDatabase()
	if err != nil {
		return nil, err
	}

	// Initialize Blockscout client
	blockscoutClient, err := client.NewBlockscoutClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Blockscout client: %w", err)
	}

	server := &Server{
		app:              app,
		database:         db,
		blockscoutClient: blockscoutClient,
	}

	// Root route - Token Management Dashboard (public, so HTML loads)
	app.Get("/", server.tokenManagementPage)

	// API routes
	api := app.Group("/api/v1")

	// Public endpoint - Token info (no authentication required)
	api.Get("/chains/:chainId/token-infos/:tokenAddress", server.getTokenInfo)

	// Protected endpoints - Token management (authentication required)
	protected := api.Group("")
	protected.Use(authMiddleware())
	{
		protected.Get("/tokens", server.getAllTokens)
		protected.Post("/tokens", server.upsertToken)
		protected.Get("/blockscout/tokens", server.getBlockscoutTokens)
		protected.Get("/blockscout/tokens/:tokenAddress", server.getBlockscoutTokenByAddress)
	}

	return server, nil
}

func (s *Server) Start(port string) error {
	return s.app.Listen(":" + port)
}

func (s *Server) Shutdown(ctx context.Context) error {
	var closeErrors []error

	// Close blockscoutClient if it exists
	if s.blockscoutClient != nil {
		if err := s.blockscoutClient.Close(); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("failed to close blockscout client: %w", err))
		}
	}

	// Close database if it exists
	if s.database != nil {
		if err := s.database.Close(); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("failed to close database: %w", err))
		}
	}

	// Shutdown the Fiber app
	if err := s.app.ShutdownWithContext(ctx); err != nil {
		closeErrors = append(closeErrors, fmt.Errorf("failed to shutdown app: %w", err))
	}

	// Return combined errors if any occurred, otherwise return nil
	if len(closeErrors) > 0 {
		return errors.Join(closeErrors...)
	}
	return nil
}

// getTokenInfo handles the token info endpoint
func (s *Server) getTokenInfo(c *fiber.Ctx) error {
	chainId := c.Params("chainId")
	tokenAddress := c.Params("tokenAddress")

	// Normalize address to lowercase to match stored format
	tokenAddress = strings.ToLower(tokenAddress)

	// Try to get token from database
	token, err := s.database.GetTokenInfo(tokenAddress, chainId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve token info",
		})
	}

	if token != nil {
		// Create a clean response structure that handles null values properly
		response := map[string]interface{}{
			"tokenAddress":        token.TokenAddress,
			"chainId":             token.ChainID,
			"projectName":         token.ProjectName,
			"projectWebsite":      token.ProjectWebsite,
			"projectEmail":        token.ProjectEmail,
			"iconUrl":             token.IconURL,
			"projectDescription":  token.ProjectDescription,
			"projectSector":       getStringValue(token.ProjectSector),
			"docs":                getStringValue(token.Docs),
			"github":              token.Github,
			"telegram":            token.Telegram,
			"linkedin":            token.Linkedin,
			"discord":             token.Discord,
			"slack":               token.Slack,
			"twitter":             token.Twitter,
			"openSea":             getStringValue(token.OpenSea),
			"facebook":            token.Facebook,
			"medium":              token.Medium,
			"reddit":              token.Reddit,
			"support":             token.Support,
			"coinMarketCapTicker": token.CoinMarketCapTicker,
			"coinGeckoTicker":     token.CoinGeckoTicker,
			"defiLlamaTicker":     token.DefiLlamaTicker,
			"tokenName":           token.TokenName,
			"tokenSymbol":         token.TokenSymbol,
		}
		return c.JSON(response)
	}

	// Return empty structure if token not found in sidecar database
	emptyToken := map[string]interface{}{
		"chainId":             "0",
		"projectName":         "",
		"projectWebsite":      "",
		"projectEmail":        "",
		"iconUrl":             "",
		"projectDescription":  "",
		"projectSector":       "",
		"docs":                "",
		"github":              "",
		"telegram":            "",
		"linkedin":            "",
		"discord":             "",
		"slack":               "",
		"twitter":             "",
		"openSea":             "",
		"facebook":            "",
		"medium":              "",
		"reddit":              "",
		"support":             "",
		"coinMarketCapTicker": "",
		"coinGeckoTicker":     "",
		"defiLlamaTicker":     "",
		"tokenAddress":        "",
		"tokenName":           "",
		"tokenSymbol":         "",
	}

	return c.JSON(emptyToken)
}

// upsertToken creates or updates token information using PostgreSQL upsert
func (s *Server) upsertToken(c *fiber.Ctx) error {
	var form models.TokenInfoForm
	if err := c.BodyParser(&form); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if form.TokenAddress == "" || form.ChainID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Token address and chain ID are required",
		})
	}

	// Normalize token address (lowercase)
	form.TokenAddress = strings.ToLower(form.TokenAddress)

	// Use the database upsert function
	err := s.database.UpsertTokenInfo(&form)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save/update token info",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Token saved/updated successfully",
	})
}

// tokenManagementPage serves the HTML page for token management
func (s *Server) tokenManagementPage(c *fiber.Ctx) error {
	// Get the embedded HTML template content
	htmlContent, err := getTemplateContent()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Template not found")
	}

	c.Set("Content-Type", "text/html")
	return c.SendString(htmlContent)
}

// getAllTokens returns all tokens
func (s *Server) getAllTokens(c *fiber.Ctx) error {
	tokens, err := s.database.GetAllTokens()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve tokens",
		})
	}

	return c.JSON(fiber.Map{
		"tokens": tokens,
		"total":  len(tokens),
	})
}

// getBlockscoutTokens fetches all tokens from Blockscout
func (s *Server) getBlockscoutTokens(c *fiber.Ctx) error {
	tokens, err := s.blockscoutClient.GetTokens()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch tokens from Blockscout",
		})
	}

	return c.JSON(fiber.Map{
		"tokens": tokens,
		"total":  len(tokens),
	})
}

// getBlockscoutTokenByAddress fetches a specific token from Blockscout by address
func (s *Server) getBlockscoutTokenByAddress(c *fiber.Ctx) error {
	tokenAddress := c.Params("tokenAddress")

	token, err := s.blockscoutClient.GetTokenByAddress(tokenAddress)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch token from Blockscout",
		})
	}

	if token == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Token not found",
		})
	}

	return c.JSON(token)
}
