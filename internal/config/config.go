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

	// Logger configuration
	// Available: "trace", "debug", "info", "warn", "error", "fatal", "panic"
	// ref: https://pkg.go.dev/github.com/rs/zerolog@v1.26.1#pkg-variables
	config.SetDefault("LOGGER_LOG_LEVEL", "info")
	config.SetDefault("LOGGER_DURATION_FIELD_UNIT", "ms") // Available: "ms", "millisecond", "s", "second"
	config.SetDefault("LOGGER_FORMAT", "json")            // Available: "json", "console"

	// Prometheus configuration
	config.SetDefault("PROMETHEUS", true)
	config.SetDefault("PROMETHEUS_PATH", "/metrics")

	// pprof configuration
	config.SetDefault("PPROF", false)

	// Redis configuration
	config.SetDefault("REDIS_CONNECTION_STRING", "redis://localhost:6379")
	config.SetDefault("REDIS_KEY_TTL", 24*time.Hour)

	// Set default IP Geolocation API
	config.SetDefault("GEOLOCATION_API", "ip-api") // Available: "ipapi", "ipbase"

	// Configuration for ip-api.com API
	config.SetDefault("IP_API_BASE_URL", "http://ip-api.com/json/") // https isn't available for free usage

	// Configuration for ipbase.com API
	config.SetDefault("IPBASE_BASE_URL", "https://api.ipbase.com/v2/info/?ip=")
	config.SetDefault("IPBASE_API_KEY", "")

	// Set default http client configuration
	config.SetDefault("HTTP_CLIENT_TIMEOUT", 15*time.Second)
}
