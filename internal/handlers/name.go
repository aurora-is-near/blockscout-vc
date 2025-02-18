package handlers

import (
	"blockscout-vc/internal/docker"
	"fmt"

	"github.com/spf13/viper"
)

// MaxCoinLength defines the maximum allowed length for a coin symbol
const MaxNameLength = 40

type NameHandler struct {
	BaseHandler
}

func NewNameHandler() *NameHandler {
	return &NameHandler{
		BaseHandler: NewBaseHandler(),
	}
}

// Handle processes coin-related changes and updates service configurations
func (h *NameHandler) Handle(record *Record) HandlerResult {
	result := HandlerResult{}

	if err := h.validateName(record.Name); err != nil {
		result.Error = fmt.Errorf("invalid name: %w", err)
		return result
	}

	frontendServiceName := viper.GetString("frontendServiceName")
	frontendContainerName := viper.GetString("frontendContainerName")

	// Create updates with string values
	updates := map[string]map[string]string{
		frontendServiceName: {
			"NEXT_PUBLIC_NETWORK_NAME":       record.Name,
			"NEXT_PUBLIC_NETWORK_SHORT_NAME": record.Name,
		},
	}

	// Apply updates to services
	for service, env := range updates {
		updated, err := h.UpdateServiceEnv(service, env)
		if err != nil {
			result.Error = fmt.Errorf("failed to update %s service environment: %w", service, err)
			return result
		}
		if updated {
			fmt.Printf("Updated %s service environment: %+v\n", service, env)
			fmt.Printf("Frontend container name: %s\n", frontendContainerName)
			fmt.Printf("Frontend service name: %s\n", frontendServiceName)
			result.ContainersToRestart = append(result.ContainersToRestart, docker.Container{
				Name:        frontendContainerName,
				ServiceName: frontendServiceName,
			})
		}
	}

	return result
}

// validateCoin checks if the coin symbol meets the required criteria
func (h *NameHandler) validateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}
	if len(name) > MaxNameLength {
		return fmt.Errorf("name length cannot exceed %d characters", MaxNameLength)
	}
	return nil
}
