package models

import (
	"database/sql"
)

// TokenInfo represents the token information structure
type TokenInfo struct {
	TokenAddress        string         `json:"tokenAddress" db:"token_address"`
	ChainID             string         `json:"chainId" db:"chain_id"`
	ProjectName         string         `json:"projectName" db:"project_name"`
	ProjectWebsite      string         `json:"projectWebsite" db:"project_website"`
	ProjectEmail        string         `json:"projectEmail" db:"project_email"`
	IconURL             string         `json:"iconUrl" db:"icon_url"`
	ProjectDescription  string         `json:"projectDescription" db:"project_description"`
	ProjectSector       sql.NullString `json:"projectSector" db:"project_sector"`
	Docs                sql.NullString `json:"docs" db:"docs"`
	Github              string         `json:"github" db:"github"`
	Telegram            string         `json:"telegram" db:"telegram"`
	Linkedin            string         `json:"linkedin" db:"linkedin"`
	Discord             string         `json:"discord" db:"discord"`
	Slack               string         `json:"slack" db:"slack"`
	Twitter             string         `json:"twitter" db:"twitter"`
	OpenSea             sql.NullString `json:"openSea" db:"opensea"`
	Facebook            string         `json:"facebook" db:"facebook"`
	Medium              string         `json:"medium" db:"medium"`
	Reddit              string         `json:"reddit" db:"reddit"`
	Support             string         `json:"support" db:"support"`
	CoinMarketCapTicker string         `json:"coinMarketCapTicker" db:"coin_market_cap_ticker"`
	CoinGeckoTicker     string         `json:"coinGeckoTicker" db:"coin_gecko_ticker"`
	DefiLlamaTicker     string         `json:"defiLlamaTicker" db:"defi_llama_ticker"`
	TokenName           string         `json:"tokenName" db:"token_name"`
	TokenSymbol         string         `json:"tokenSymbol" db:"token_symbol"`
}

// TokenInfoForm represents the form data for creating/updating tokens
type TokenInfoForm struct {
	TokenAddress        string `json:"tokenAddress" form:"tokenAddress"`
	ChainID             string `json:"chainId" form:"chainId"`
	ProjectName         string `json:"projectName" form:"projectName"`
	ProjectWebsite      string `json:"projectWebsite" form:"projectWebsite"`
	ProjectEmail        string `json:"projectEmail" form:"projectEmail"`
	IconURL             string `json:"iconUrl" form:"iconUrl"`
	ProjectDescription  string `json:"projectDescription" form:"projectDescription"`
	ProjectSector       string `json:"projectSector" form:"projectSector"`
	Docs                string `json:"docs" form:"docs"`
	Github              string `json:"github" form:"github"`
	Telegram            string `json:"telegram" form:"telegram"`
	Linkedin            string `json:"linkedin" form:"linkedin"`
	Discord             string `json:"discord" form:"discord"`
	Slack               string `json:"slack" form:"slack"`
	Twitter             string `json:"twitter" form:"twitter"`
	OpenSea             string `json:"openSea" form:"openSea"`
	Facebook            string `json:"facebook" form:"facebook"`
	Medium              string `json:"medium" form:"medium"`
	Reddit              string `json:"reddit" form:"reddit"`
	Support             string `json:"support" form:"support"`
	CoinMarketCapTicker string `json:"coinMarketCapTicker" form:"coinMarketCapTicker"`
	CoinGeckoTicker     string `json:"coinGeckoTicker" form:"coinGeckoTicker"`
	DefiLlamaTicker     string `json:"defiLlamaTicker" form:"defiLlamaTicker"`
	TokenName           string `json:"tokenName" form:"tokenName"`
	TokenSymbol         string `json:"tokenSymbol" form:"tokenSymbol"`
}
