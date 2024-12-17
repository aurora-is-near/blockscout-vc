package config

import (
	"log"

	"github.com/spf13/viper"
)

// InitConfig initializes the application configuration using viper
// It looks for a configuration file named 'local.yaml' in the config directory
func InitConfig() {
	viper.AddConfigPath("config") // path to look for the config file in
	viper.SetConfigName("local")  // name of the config file (without extension)
	viper.SetConfigType("yaml")   // type of the config file

	// Enable automatic environment variable binding
	viper.AutomaticEnv()

	// Attempt to read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Can't read config: %s\n", err)
	}
}
