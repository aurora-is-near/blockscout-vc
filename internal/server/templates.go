package server

import "embed"

//go:embed templates/*.html
var embedTemplates embed.FS

// getTemplateContent returns the embedded HTML template content
func getTemplateContent() (string, error) {
	content, err := embedTemplates.ReadFile("templates/token-management.html")
	if err != nil {
		return "", err
	}
	return string(content), nil
}
