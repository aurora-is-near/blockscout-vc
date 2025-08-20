package client

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

// BlockscoutClient represents a client for interacting with Blockscout database
type BlockscoutClient struct {
	db *sql.DB
}

// BlockscoutToken represents a token from Blockscout database
type BlockscoutToken struct {
	Address string `json:"address"`
	Symbol  string `json:"symbol"`
	Name    string `json:"name"`
}

// NewBlockscoutClient creates a new Blockscout client with direct database access
func NewBlockscoutClient() (*BlockscoutClient, error) {
	// Get database connection string from config
	databaseURL := viper.GetString("blockscoutDatabaseUrl")
	if databaseURL == "" {
		return nil, fmt.Errorf("blockscoutDatabaseUrl not configured")
	}

	// Open database connection
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open blockscout database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping blockscout database: %w", err)
	}

	return &BlockscoutClient{db: db}, nil
}

// Close closes the database connection
func (c *BlockscoutClient) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// GetTokens fetches all tokens from Blockscout database
func (c *BlockscoutClient) GetTokens() ([]BlockscoutToken, error) {
	// Get all tokens
	query := `
		SELECT regexp_replace(contract_address_hash::varchar, '^\\x', '0x'), symbol, name
		FROM tokens
		ORDER BY name ASC
	`

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tokens: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close rows: %v\n", closeErr)
		}
	}()

	var tokens []BlockscoutToken
	for rows.Next() {
		var token BlockscoutToken
		err := rows.Scan(
			&token.Address,
			&token.Symbol,
			&token.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan token: %w", err)
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

// GetTokenByAddress fetches a specific token from Blockscout database by address
func (c *BlockscoutClient) GetTokenByAddress(address string) (*BlockscoutToken, error) {
	query := `
		SELECT regexp_replace(contract_address_hash::varchar, '^\\x', '0x'), symbol, name
		FROM tokens
		WHERE regexp_replace(contract_address_hash::varchar, '^\\x', '0x') = $1
	`

	var token BlockscoutToken
	err := c.db.QueryRow(query, address).Scan(
		&token.Address,
		&token.Symbol,
		&token.Name,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Token not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get token by address: %w", err)
	}

	return &token, nil
}
