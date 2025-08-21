package client

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

// BlockscoutClient represents a client for interacting with Blockscout database
// Note: COALESCE is used for symbol and name fields as they can be NULL in the Blockscout database schema.
// Contract address matching uses case-insensitive comparison for better user experience.
type BlockscoutClient struct {
	db *sql.DB
}

// BlockscoutToken represents a token from Blockscout database
type BlockscoutToken struct {
	Address string `json:"address"`
	Symbol  string `json:"symbol"`
	Name    string `json:"name"`
	IconURL string `json:"icon_url"`
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
	// Get all tokens - use COALESCE to handle NULL values for symbol, name, and icon_url
	query := `
		SELECT regexp_replace(contract_address_hash::varchar, '^\\x', '0x'), 
		       COALESCE(symbol, '') as symbol, 
		       COALESCE(name, '') as name,
		       COALESCE(icon_url, '') as icon_url
		FROM tokens
		ORDER BY COALESCE(name, '') ASC
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
			&token.IconURL,
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

// GetTokenByAddress fetches a specific token from Blockscout database by address
func (c *BlockscoutClient) GetTokenByAddress(address string) (*BlockscoutToken, error) {
	// Use COALESCE to handle NULL values for symbol, name, and icon_url
	// Use case-insensitive comparison for contract address matching
	query := `
		SELECT regexp_replace(contract_address_hash::varchar, '^\\x', '0x'), 
		       COALESCE(symbol, '') as symbol, 
		       COALESCE(name, '') as name,
		       COALESCE(icon_url, '') as icon_url
		FROM tokens
		WHERE lower(regexp_replace(contract_address_hash::varchar, '^\\x', '0x')) = lower($1)
	`

	var token BlockscoutToken
	err := c.db.QueryRow(query, address).Scan(
		&token.Address,
		&token.Symbol,
		&token.Name,
		&token.IconURL,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Token not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get token by address: %w", err)
	}

	return &token, nil
}

// UpdateTokenIconURL updates the icon_url field for a specific token in Blockscout database
func (c *BlockscoutClient) UpdateTokenIconURL(address, iconURL string) error {
	// Use case-insensitive comparison for contract address matching
	query := `
		UPDATE tokens 
		SET icon_url = $2, updated_at = CURRENT_TIMESTAMP
		WHERE lower(regexp_replace(contract_address_hash::varchar, '^\\x', '0x')) = lower($1)
	`

	result, err := c.db.Exec(query, address, iconURL)
	if err != nil {
		return fmt.Errorf("failed to update token icon_url: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no token found with address: %s", address)
	}

	return nil
}
