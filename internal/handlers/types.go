package handlers

import "blockscout-vc/internal/docker"

// Handler defines the interface for all update handlers
type Handler interface {
	Handle(record *Record) HandlerResult
}

// HandlerResult represents the outcome of a handler's processing
type HandlerResult struct {
	Error               error              // Any error that occurred during handling
	ContainersToRestart []docker.Container // List of container names that need to be restarted
}

// Record represents the common data structure for all handlers
// containing the database record fields
type Record struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Coin         string `json:"base_token_symbol"`
	ChainID      string `json:"chain_id"`
	LightLogoURL string `json:"network_logo"`
	DarkLogoURL  string `json:"network_logo_dark"`
	FaviconURL   string `json:"favicon"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// BaseHandler provides common functionality for handlers
type BaseHandler struct {
	docker *docker.Docker
}

func NewBaseHandler() BaseHandler {
	return BaseHandler{
		docker: docker.NewDocker(),
	}
}

type EnvUpdate struct {
	ServiceName   string
	Key           string
	Value         string
	ContainerName string
}
