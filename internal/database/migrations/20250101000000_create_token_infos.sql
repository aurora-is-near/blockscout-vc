-- +goose Up
CREATE TABLE IF NOT EXISTS token_infos (
    id SERIAL PRIMARY KEY,
    token_address VARCHAR(42) NOT NULL,
    chain_id VARCHAR(20) NOT NULL,
    project_name VARCHAR(255),
    project_website VARCHAR(255),
    project_email VARCHAR(255),
    icon_url TEXT,
    project_description TEXT,
    project_sector VARCHAR(255),
    docs VARCHAR(255),
    github VARCHAR(255),
    telegram VARCHAR(255),
    linkedin VARCHAR(255),
    discord VARCHAR(255),
    slack VARCHAR(255),
    twitter VARCHAR(255),
    opensea VARCHAR(255),
    facebook VARCHAR(255),
    medium VARCHAR(255),
    reddit VARCHAR(255),
    support TEXT,
    coin_market_cap_ticker VARCHAR(255),
    coin_gecko_ticker VARCHAR(255),
    defi_llama_ticker VARCHAR(255),
    token_name VARCHAR(255),
    token_symbol VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(token_address, chain_id)
);

CREATE INDEX IF NOT EXISTS idx_token_infos_token_address ON token_infos(token_address);
CREATE INDEX IF NOT EXISTS idx_token_infos_chain_id ON token_infos(chain_id);

-- +goose Down
DROP TABLE IF EXISTS token_infos; 