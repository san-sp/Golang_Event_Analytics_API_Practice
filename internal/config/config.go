package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config contains application settings read from the environment.
// Keeping configuration in one place avoids hardcoding deployment
// settings throughout the application.
type Config struct {
	Port          string
	MongoURI      string
	MongoDatabase string
}

// Load reads environment variables and applies development-friendly
// defaults. It rejects invalid port values before the server starts.
func Load() (Config, error) {
	cfg := Config{
		Port:          getEnv("PORT", "8100"),
		MongoURI:      getEnv("MONGODB_URI", "mongodb://localhost:27018"),
		MongoDatabase: getEnv("MONGODB_DATABASE", "pulsetrack"),
	}

	// A port must be a valid TCP port number.
	port, err := strconv.Atoi(cfg.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT must be a number between 1 and 65535")
	}

	if cfg.MongoURI == "" {
		return Config{}, fmt.Errorf("MONGODB_URI cannot be empty")
	}

	if cfg.MongoDatabase == "" {
		return Config{}, fmt.Errorf("MONGODB_DATABASE cannot be empty")
	}

	return cfg, nil
}

// getEnv returns the configured environment value or its fallback.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return fallback
}
