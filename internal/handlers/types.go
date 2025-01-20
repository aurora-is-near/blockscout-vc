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
	Coin         string `json:"coin"`
	ChainID      int    `json:"chain_id"`
	LightLogoURL string `json:"light_logo_url"`
	DarkLogoURL  string `json:"dark_logo_url"`
	FaviconURL   string `json:"favicon_url"`
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
