package handlers

import (
	"blockscout-vc/internal/docker"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// MaxImageLength defines the maximum allowed length for image URLs
const MaxImageLength = 2000

type ImageHandler struct {
	BaseHandler
	client *http.Client
}

func NewImageHandler() *ImageHandler {
	return &ImageHandler{
		BaseHandler: NewBaseHandler(),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
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

	frontendServiceName := viper.GetString("frontendServiceName")
	frontendContainerName := viper.GetString("frontendContainerName")

	// Initialize updates with string map
	updates := map[string]map[string]string{
		frontendServiceName: make(map[string]string),
	}

	// Validate and update light logo URL
	if err := h.validateImage(record.LightLogoURL); err != nil {
		result.Error = fmt.Errorf("invalid light logo URL: %w", err)
	} else {
		updates[frontendServiceName]["NEXT_PUBLIC_NETWORK_LOGO"] = record.LightLogoURL
	}

	// Validate and update dark logo URL
	if err := h.validateImage(record.DarkLogoURL); err != nil {
		result.Error = fmt.Errorf("invalid dark logo URL: %w", err)
	} else {
		updates[frontendServiceName]["NEXT_PUBLIC_NETWORK_LOGO_DARK"] = record.DarkLogoURL
	}

	// Validate and update favicon URL
	if err := h.validateImage(record.FaviconURL); err != nil {
		result.Error = fmt.Errorf("invalid favicon URL: %w", err)
	} else {
		updates[frontendServiceName]["NEXT_PUBLIC_NETWORK_ICON"] = record.FaviconURL
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

// validateImage checks if the image URL meets the required criteria
func (h *ImageHandler) validateImage(imageURL string) error {
	if imageURL == "" {
		return fmt.Errorf("image cannot be empty")
	}

	if len(imageURL) > MaxImageLength {
		return fmt.Errorf("image length cannot exceed %d characters", MaxImageLength)
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check if scheme is http or https
	if !strings.HasPrefix(parsedURL.Scheme, "http") {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	// Check if image is accessible
	resp, err := h.client.Head(imageURL)
	if err != nil {
		return fmt.Errorf("failed to access image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("image not accessible, status code: %d", resp.StatusCode)
	}

	// Optionally verify content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("URL does not point to an image (content-type: %s)", contentType)
	}

	return nil
}
