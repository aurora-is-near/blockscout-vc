-- +goose Up
-- Simplified schema: All string fields use TEXT (no length limits), 
-- natural primary key (token_address, chain_id), no unnecessary id field
CREATE TABLE IF NOT EXISTS token_infos (
    token_address TEXT NOT NULL,
    chain_id BIGINT NOT NULL,
    project_name TEXT,
    project_website TEXT,
    project_email TEXT,
    icon_url TEXT,
    project_description TEXT,
    project_sector TEXT,
    docs TEXT,
    github TEXT,
    telegram TEXT,
    linkedin TEXT,
    discord TEXT,
    slack TEXT,
    twitter TEXT,
    opensea TEXT,
    facebook TEXT,
    medium TEXT,
    reddit TEXT,
    support TEXT,
    coin_market_cap_ticker TEXT,
    coin_gecko_ticker TEXT,
    defi_llama_ticker TEXT,
    token_name TEXT,
    token_symbol TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (token_address, chain_id)
);

-- Create indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_token_infos_token_address_lower ON token_infos(lower(token_address));
CREATE INDEX IF NOT EXISTS idx_token_infos_chain_id ON token_infos(chain_id);

-- +goose Down
DROP TABLE IF EXISTS token_infos; 