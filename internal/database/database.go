package database

import (
	"blockscout-vc/internal/models"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

// Database provides database operations for token information
// Note: All string fields use TEXT type which can handle NULL values without scan errors
type Database struct {
	db *sql.DB
}

func NewDatabase() (*Database, error) {
	// Get database connection string from config
	databaseURL := viper.GetString("sidecarDatabaseUrl")
	if databaseURL == "" {
		return nil, fmt.Errorf("sidecarDatabaseUrl not configured")
	}

	// Create database if it doesn't exist
	if err := createDatabaseIfNotExists(databaseURL); err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Open database connection
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to ping database: %w, and failed to close connection: %w", err, closeErr)
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to run migrations: %w, and failed to close connection: %w", err, closeErr)
		}
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

// GetTokenInfo retrieves token information by token address and chain ID
func (d *Database) GetTokenInfo(tokenAddress, chainID string) (*models.TokenInfo, error) {
	query := `
		SELECT token_address, chain_id, project_name, project_website, project_email,
		       icon_url, project_description, project_sector, docs, github, telegram,
		       linkedin, discord, slack, twitter, opensea, facebook, medium, reddit,
		       support, coin_market_cap_ticker, coin_gecko_ticker, defi_llama_ticker,
		       token_name, token_symbol
		FROM token_infos
		WHERE token_address = $1 AND chain_id = $2
	`

	var token models.TokenInfo
	err := d.db.QueryRow(query, tokenAddress, chainID).Scan(
		&token.TokenAddress, &token.ChainID, &token.ProjectName,
		&token.ProjectWebsite, &token.ProjectEmail, &token.IconURL,
		&token.ProjectDescription, &token.ProjectSector, &token.Docs,
		&token.Github, &token.Telegram, &token.Linkedin, &token.Discord,
		&token.Slack, &token.Twitter, &token.OpenSea, &token.Facebook,
		&token.Medium, &token.Reddit, &token.Support, &token.CoinMarketCapTicker,
		&token.CoinGeckoTicker, &token.DefiLlamaTicker, &token.TokenName,
		&token.TokenSymbol,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Token not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get token info: %w", err)
	}

	return &token, nil
}

// GetAllTokens retrieves all tokens
func (d *Database) GetAllTokens() ([]models.TokenInfo, error) {
	query := `
		SELECT token_address, chain_id, project_name, project_website, project_email,
		       icon_url, project_description, project_sector, docs, github, telegram,
		       linkedin, discord, slack, twitter, opensea, facebook, medium, reddit,
		       support, coin_market_cap_ticker, coin_gecko_ticker, defi_llama_ticker,
		       token_name, token_symbol
		FROM token_infos
		ORDER BY created_at DESC
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tokens: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// Log the error but don't return it since we're in a defer
			fmt.Printf("Warning: failed to close rows: %v\n", closeErr)
		}
	}()

	var tokens []models.TokenInfo
	for rows.Next() {
		var token models.TokenInfo
		err := rows.Scan(
			&token.TokenAddress, &token.ChainID, &token.ProjectName,
			&token.ProjectWebsite, &token.ProjectEmail, &token.IconURL,
			&token.ProjectDescription, &token.ProjectSector, &token.Docs,
			&token.Github, &token.Telegram, &token.Linkedin, &token.Discord,
			&token.Slack, &token.Twitter, &token.OpenSea, &token.Facebook,
			&token.Medium, &token.Reddit, &token.Support, &token.CoinMarketCapTicker,
			&token.CoinGeckoTicker, &token.DefiLlamaTicker, &token.TokenName,
			&token.TokenSymbol,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan token: %w", err)
		}
		tokens = append(tokens, token)
	}

	// Check for errors from iterating over rows
	// This is crucial: rows.Err() catches errors that might occur during iteration
	// that aren't caught by the individual rows.Scan() calls
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return tokens, nil
}

// UpsertTokenInfo creates or updates token information using PostgreSQL upsert
// Manually sets updated_at timestamp instead of relying on database triggers
func (d *Database) UpsertTokenInfo(form *models.TokenInfoForm) error {
	query := `
		INSERT INTO token_infos (
			token_address, chain_id, project_name, project_website, project_email,
			icon_url, project_description, project_sector, docs, github, telegram,
			linkedin, discord, slack, twitter, opensea, facebook, medium, reddit,
			support, coin_market_cap_ticker, coin_gecko_ticker, defi_llama_ticker,
			token_name, token_symbol, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, CURRENT_TIMESTAMP
		)
		ON CONFLICT ON CONSTRAINT token_infos_pkey
		DO UPDATE SET
			project_name = EXCLUDED.project_name,
			project_website = EXCLUDED.project_website,
			project_email = EXCLUDED.project_email,
			icon_url = EXCLUDED.icon_url,
			project_description = EXCLUDED.project_description,
			project_sector = EXCLUDED.project_sector,
			docs = EXCLUDED.docs,
			github = EXCLUDED.github,
			telegram = EXCLUDED.telegram,
			linkedin = EXCLUDED.linkedin,
			discord = EXCLUDED.discord,
			slack = EXCLUDED.slack,
			twitter = EXCLUDED.twitter,
			opensea = EXCLUDED.opensea,
			facebook = EXCLUDED.facebook,
			medium = EXCLUDED.medium,
			reddit = EXCLUDED.reddit,
			support = EXCLUDED.support,
			coin_market_cap_ticker = EXCLUDED.coin_market_cap_ticker,
			coin_gecko_ticker = EXCLUDED.coin_gecko_ticker,
			defi_llama_ticker = EXCLUDED.defi_llama_ticker,
			token_name = EXCLUDED.token_name,
			token_symbol = EXCLUDED.token_symbol,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := d.db.Exec(query,
		form.TokenAddress, form.ChainID, form.ProjectName, form.ProjectWebsite,
		form.ProjectEmail, form.IconURL, form.ProjectDescription, form.ProjectSector,
		form.Docs, form.Github, form.Telegram, form.Linkedin, form.Discord,
		form.Slack, form.Twitter, form.OpenSea, form.Facebook, form.Medium,
		form.Reddit, form.Support, form.CoinMarketCapTicker, form.CoinGeckoTicker,
		form.DefiLlamaTicker, form.TokenName, form.TokenSymbol,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert token info: %w", err)
	}

	log.Printf("Upserted token: %s on chain %s", form.TokenAddress, form.ChainID)
	return nil
}
