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

	// Validate and update light logo URL
	if err := h.validateImage(record.LightLogoURL); err != nil {
		result.Error = fmt.Errorf("invalid light logo URL: %w", err)
		return result
	}

	// Validate and update dark logo URL
	if err := h.validateImage(record.DarkLogoURL); err != nil {
		result.Error = fmt.Errorf("invalid dark logo URL: %w", err)
		return result
	}

	// Validate and update favicon URL
	if err := h.validateImage(record.FaviconURL); err != nil {
		result.Error = fmt.Errorf("invalid favicon URL: %w", err)
		return result
	}

	compose, err := h.docker.ReadComposeFile()
	if err != nil {
		result.Error = fmt.Errorf("failed to read compose file: %w", err)
		return result
	}

	frontendServiceName := viper.GetString("frontendServiceName")
	frontendContainerName := viper.GetString("frontendContainerName")

	updates := []docker.EnvUpdate{
		{
			ServiceName:   frontendServiceName,
			Key:           "NEXT_PUBLIC_NETWORK_LOGO",
			Value:         record.LightLogoURL,
			ContainerName: frontendContainerName,
		},
		{
			ServiceName:   frontendServiceName,
			Key:           "NEXT_PUBLIC_NETWORK_LOGO_DARK",
			Value:         record.DarkLogoURL,
			ContainerName: frontendContainerName,
		},
		{
			ServiceName:   frontendServiceName,
			Key:           "NEXT_PUBLIC_NETWORK_ICON",
			Value:         record.FaviconURL,
			ContainerName: frontendContainerName,
		},
	}

	result.ContainersToRestart, err = h.docker.AppyAndUpdateEachService(compose, updates)
	if err != nil {
		result.Error = err
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
