package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	CORS CORSConfig
	Auth AuthConfig
}

// CORSConfig holds CORS-related configuration
type CORSConfig struct {
	AllowedOrigins []string
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	Username string
	Password string
}

// GetCORSAllowedOrigins returns the list of allowed CORS origins
func GetCORSAllowedOrigins() []string {
	originsStr := viper.GetString("cors.allowedOrigins")
	if originsStr == "" {
		// Safe default: no origins allowed
		return []string{}
	}

	// Split comma-separated origins and trim whitespace
	origins := strings.Split(originsStr, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}

	return origins
}

// GetAuthUsername returns the authentication username
func GetAuthUsername() string {
	return viper.GetString("auth.username")
}

// GetAuthPassword returns the authentication password
func GetAuthPassword() string {
	return viper.GetString("auth.password")
}

// InitConfig initializes the application configuration using viper.
// If configPath is provided, it will use that specific file,
// otherwise it will look for 'local.yaml' in the config directory
func InitConfig(configPath string) {
	if configPath != "" {
		// Use specified config file
		viper.SetConfigFile(configPath)
	} else {
		// Default config location
		viper.AddConfigPath("config")
		viper.SetConfigName("local")
	}
	viper.SetConfigType("yaml")

	// Enable automatic environment variable binding
	viper.AutomaticEnv()

	// Attempt to read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Can't read config: %s\n", err)
	}
}
