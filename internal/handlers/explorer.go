package handlers

import (
	"blockscout-vc/internal/docker"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

// MaxExplorerURLLength defines the maximum allowed length for an explorer URL
const MaxExplorerURLLength = 255

type ExplorerHandler struct {
	BaseHandler
}

func NewExplorerHandler() *ExplorerHandler {
	return &ExplorerHandler{
		BaseHandler: NewBaseHandler(),
	}
}

// Handle processes explorer URL changes and updates service configurations
func (h *ExplorerHandler) Handle(record *Record) HandlerResult {
	result := HandlerResult{}

	if err := h.validateExplorerURL(record.ExplorerURL); err != nil {
		result.Error = fmt.Errorf("invalid explorer URL: %w", err)
		return result
	}

	// Extract host from explorer URL
	host, err := h.extractHostFromURL(record.ExplorerURL)
	if err != nil {
		result.Error = fmt.Errorf("failed to extract host from explorer URL: %w", err)
		return result
	}

	// Extract protocol from explorer URL
	protocol := h.extractProtocolFromURL(record.ExplorerURL)

	// Get service names from config with defaults for backward compatibility
	frontendServiceName := viper.GetString("frontendServiceName")
	frontendContainerName := viper.GetString("frontendContainerName")
	backendServiceName := viper.GetString("backendServiceName")
	backendContainerName := viper.GetString("backendContainerName")
	statsServiceName := viper.GetString("statsServiceName")
	statsContainerName := viper.GetString("statsContainerName")

	// Get proxy service configuration - only restart if both are present
	proxyServiceName := viper.GetString("proxyServiceName")
	proxyContainerName := viper.GetString("proxyContainerName")

	// Update the sidecar-injected.env file with all explorer-related environment variables
	// This file is loaded by all services and will override values from other env files
	sidecarUpdates := map[string]string{
		"BLOCKSCOUT_HOST":                    host,
		"MICROSERVICE_VISUALIZE_SOL2UML_URL": fmt.Sprintf("%s://visualize.%s", protocol, host),
		"NEXT_PUBLIC_FEATURED_NETWORKS":      fmt.Sprintf(`[{'title':'Aurora','url':'https://explorer.aurora.dev/','group':'Mainnets'}, {'title':'%s','url':'%s://%s','group':'Mainnets', 'isActive':true}]`, record.Name, protocol, host),
		"NEXT_PUBLIC_API_HOST":               host,
		"NEXT_PUBLIC_APP_HOST":               host,
		"NEXT_PUBLIC_STATS_API_HOST":         fmt.Sprintf("%s://%s", protocol, host),
		"NEXT_PUBLIC_VISUALIZE_API_HOST":     fmt.Sprintf("%s://%s", protocol, host),
		"STATS__BLOCKSCOUT_API_URL":          fmt.Sprintf("%s://%s", protocol, host),
		"EXPLORER_URL":                       host,
		"BLOCKSCOUT_HTTP_PROTOCOL":           protocol,
	}

	// Apply updates to the sidecar-injected.env file
	updated, err := h.UpdateEnvFile(sidecarUpdates)
	if err != nil {
		result.Error = fmt.Errorf("failed to update sidecar-injected environment: %w", err)
		return result
	}

	// If any environment variables were updated, restart all services
	containersToRestart := []docker.Container{}
	if updated {
		fmt.Printf("Updated explorer host to: %s\n", host)

		containersToRestart = []docker.Container{
			{
				Name:        backendContainerName,
				ServiceName: backendServiceName,
			},
			{
				Name:        frontendContainerName,
				ServiceName: frontendServiceName,
			},
			{
				Name:        statsContainerName,
				ServiceName: statsServiceName,
			},
		}

		// Only add proxy container if both service name and container name are configured
		if proxyServiceName != "" && proxyContainerName != "" {
			containersToRestart = append(containersToRestart, docker.Container{
				Name:        proxyContainerName,
				ServiceName: proxyServiceName,
			})
		}
	}

	result.ContainersToRestart = containersToRestart
	return result
}

// validateExplorerURL checks if the explorer URL meets the required criteria
func (h *ExplorerHandler) validateExplorerURL(explorerURL string) error {
	if explorerURL == "" {
		return fmt.Errorf("explorer URL cannot be empty")
	}
	if len(explorerURL) > MaxExplorerURLLength {
		return fmt.Errorf("explorer URL length cannot exceed %d characters", MaxExplorerURLLength)
	}

	// Validate URL format
	parsedURL, err := url.Parse(explorerURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}
	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must include scheme (http:// or https://)")
	}
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must include a valid host")
	}

	return nil
}

// extractHostFromURL extracts the host from a URL string
func (h *ExplorerHandler) extractHostFromURL(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	host := parsedURL.Host
	// Remove port if present
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	return host, nil
}

// extractProtocolFromURL extracts the protocol (http or https) from a URL string
func (h *ExplorerHandler) extractProtocolFromURL(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "https" // Default to https if parsing fails
	}

	if parsedURL.Scheme == "http" {
		return "http"
	}
	return "https" // Default to https if no scheme is found
}
