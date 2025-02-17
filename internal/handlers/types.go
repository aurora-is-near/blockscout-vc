package handlers

import (
	"blockscout-vc/internal/docker"
	"blockscout-vc/internal/env"
	"fmt"
)

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
	ChainID      int    `json:"chain_id"`
	LightLogoURL string `json:"network_logo"`
	DarkLogoURL  string `json:"network_logo_dark"`
	FaviconURL   string `json:"favicon"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// BaseHandler provides common functionality for handlers
type BaseHandler struct {
	docker *docker.Docker
	env    *env.Env
}

func NewBaseHandler() BaseHandler {
	return BaseHandler{
		docker: docker.NewDocker(),
		env:    env.NewEnv(),
	}
}

type EnvUpdate struct {
	ServiceName   string
	Key           string
	Value         string
	ContainerName string
}

func (h *BaseHandler) UpdateServiceEnv(serviceName string, envVars map[string]string) (bool, error) {
	if h.env.PathToEnvFile == "" {
		err := h.docker.ReadComposeFile()
		if err != nil {
			return false, fmt.Errorf("failed to read compose file: %w", err)
		}
		updated, err := h.docker.UpdateServiceEnv(serviceName, envVars)
		if err != nil {
			return false, fmt.Errorf("failed to update service env: %w", err)
		}
		if updated {
			h.docker.WriteComposeFile()
		}
		return updated, nil
	} else {
		err := h.env.ReadEnvFile()
		if err != nil {
			return false, fmt.Errorf("failed to read env file: %w", err)
		}
		updated, err := h.env.UpdateEnvVars(envVars)
		if err != nil {
			return false, fmt.Errorf("failed to update env vars: %w", err)
		}
		if updated {
			h.env.WriteEnvFile()
		}
		return updated, nil
	}
}

func (h *BaseHandler) SaveFile() error {
	if h.env.PathToEnvFile == "" {
		return h.docker.WriteComposeFile()
	} else {
		return h.env.WriteEnvFile()
	}
}
