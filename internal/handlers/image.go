package handlers

import (
	"fmt"

	"github.com/spf13/viper"
)

// MaxImageLength defines the maximum allowed length for image URLs
const MaxImageLength = 2000

type ImageHandler struct {
	BaseHandler
}

func NewImageHandler() *ImageHandler {
	return &ImageHandler{
		BaseHandler: NewBaseHandler(),
	}
}

// Handle processes image-related changes and updates service configurations
// It handles light logo, dark logo, and favicon URL updates
func (h *ImageHandler) Handle(record *Record) HandlerResult {
	result := HandlerResult{}

	// Skip if no image URLs are provided
	if record.LightLogoURL == "" && record.DarkLogoURL == "" && record.FaviconURL == "" {
		return result
	}

	compose, err := h.docker.ReadComposeFile()
	if err != nil {
		result.Error = fmt.Errorf("failed to read compose file: %w", err)
		return result
	}

	frontendService := viper.GetString("frontendServiceName")

	updates := map[string]map[string]interface{}{
		frontendService: {},
	}

	// Validate and update light logo URL
	if err := h.validateImage(record.LightLogoURL); err != nil {
		result.Error = fmt.Errorf("invalid light logo URL: %w", err)
		return result
	} else {
		updates[frontendService]["NEXT_PUBLIC_NETWORK_LOGO"] = record.LightLogoURL
	}

	// Validate and update dark logo URL
	if err := h.validateImage(record.DarkLogoURL); err != nil {
		result.Error = fmt.Errorf("invalid dark logo URL: %w", err)
		return result
	} else {
		updates[frontendService]["NEXT_PUBLIC_NETWORK_LOGO_DARK"] = record.DarkLogoURL
	}

	// Validate and update favicon URL
	if err := h.validateImage(record.FaviconURL); err != nil {
		result.Error = fmt.Errorf("invalid favicon URL: %w", err)
		return result
	} else {
		updates[frontendService]["NEXT_PUBLIC_NETWORK_ICON"] = record.FaviconURL
	}

	// Apply updates to services
	for service, env := range updates {
		var updated bool
		compose, updated, err = h.docker.UpdateServiceEnv(compose, service, env)
		if err != nil {
			result.Error = fmt.Errorf("failed to update %s service environment: %w", service, err)
			return result
		}
		if updated {
			fmt.Printf("Updated %s service environment: %+v\n", service, env)
			result.ContainersToRestart = append(result.ContainersToRestart, service)
		}
	}

	err = h.docker.WriteComposeFile(compose)
	if err != nil {
		result.Error = fmt.Errorf("failed to write compose file: %w", err)
		return result
	}

	return result
}

// validateImage checks if the image URL meets the required criteria
func (h *ImageHandler) validateImage(image string) error {
	if image == "" {
		return fmt.Errorf("image cannot be empty")
	}
	if len(image) == 0 {
		return fmt.Errorf("image cannot be empty")
	}
	if len(image) > MaxImageLength {
		return fmt.Errorf("image length cannot exceed %d characters", MaxImageLength)
	}
	return nil
}
