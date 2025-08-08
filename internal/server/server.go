package server

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Server struct {
	app *fiber.App
}

func NewServer() *Server {
	app := fiber.New(fiber.Config{
		AppName: "Blockscout VC API",
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))
	// API routes
	api := app.Group("/api/v1")

	// Token info endpoint
	api.Get("/chains/:chainId/token-infos/:tokenAddress", getTokenInfo)

	return &Server{
		app: app,
	}
}

func (s *Server) Start(port string) error {
	return s.app.Listen(":" + port)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}

// getTokenInfo handles the token info endpoint
func getTokenInfo(c *fiber.Ctx) error {
	chainId := c.Params("chainId")
	tokenAddress := c.Params("tokenAddress")

	// For now, return hardcoded JSON for the specific token
	// In the future, this could be fetched from a database or external API
	if chainId == "1313161567" && tokenAddress == "0x80Da25Da4D783E57d2FCdA0436873A193a4BEccF" {
		tokenInfo := map[string]interface{}{
			"tokenAddress":        "0x80da25da4d783e57d2fcda0436873a193a4beccf",
			"chainId":             "1313161567",
			"projectName":         "Bridged USDT Aurora",
			"projectWebsite":      "https://tether.to",
			"projectEmail":        "support@tether.to",
			"iconUrl":             "https://assets.coingecko.com/coins/images/35055/small/USDT.png?1707233747",
			"projectDescription":  "Tether tokens are the most widely adopted stablecoins, having pioneered the concept in the digital token space.(Sidecar)",
			"projectSector":       nil,
			"docs":                nil,
			"github":              "https://github.com",
			"telegram":            "https://t.me",
			"linkedin":            "https://linkedin.com",
			"discord":             "https://discord",
			"slack":               "https://slack.com",
			"twitter":             "https://x.com",
			"openSea":             nil,
			"facebook":            "https://fb.com",
			"medium":              "https://medium.com",
			"reddit":              "https://redit.com",
			"support":             "",
			"coinMarketCapTicker": "https://coin-market.com",
			"coinGeckoTicker":     "https://coingecko.com",
			"defiLlamaTicker":     "",
			"tokenName":           "TetherAdmin",
			"tokenSymbol":         "TETH",
		}

		return c.JSON(tokenInfo)
	}

	// Return 404 for other tokens
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "Token not found",
	})
}
