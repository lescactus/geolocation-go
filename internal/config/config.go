package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

const (
	// Name of the application
	AppName = "geolocation-go"
)

type Config struct {
	*viper.Viper
}

func New() *Config {
	config := &Config{
		Viper: viper.New(),
	}

	// Set default configurations
	config.setDefaults()

	// Select the .env file
	config.SetConfigName(config.GetString("APP_CONFIG_NAME"))
	config.SetConfigType("dotenv")
	config.AddConfigPath(config.GetString("APP_CONFIG_PATH"))

	// Automatically refresh environment variables
	config.AutomaticEnv()

	// Read configuration
	if err := config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Println("failed to read configuration:", err.Error())
			os.Exit(1)
		}
	}

	return config
}

func (config *Config) setDefaults() {
	// Set default App configuration
	config.SetDefault("APP_ADDR", ":8080")
	config.SetDefault("APP_CONFIG_NAME", ".env")
	config.SetDefault("APP_CONFIG_PATH", ".")

	// Server configuration
	config.SetDefault("SERVER_READ_TIMEOUT", 30*time.Second)
	config.SetDefault("SERVER_READ_HEADER_TIMEOUT", 10*time.Second)
	config.SetDefault("SERVER_WRITE_TIMEOUT", 30*time.Second)

	// pprof configuration
	config.SetDefault("PPROF", false)

	// Redis configuration
	config.SetDefault("REDIS_CONNECTION_STRING", "redis://localhost:6379")

	// Set default IP Geolocation API
	config.SetDefault("GEOLOCATION_API", "ip-api") // Available: "ipapi"

	// Configuration for ip-api.com API
	config.SetDefault("IP_API_BASE_URL", "http://ip-api.com/json/") // https isn't available for free usage

	// Set default http client configuration
	config.SetDefault("HTTP_CLIENT_TIMEOUT", 15*time.Second)
}
