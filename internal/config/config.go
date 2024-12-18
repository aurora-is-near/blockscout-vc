package config

import (
	"log"

	"github.com/spf13/viper"
)

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
