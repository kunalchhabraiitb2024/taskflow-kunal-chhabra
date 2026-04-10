package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	ServerPort  string
	BcryptCost  int
}

// Load reads configuration from environment variables.
// Panics on startup if required values are missing — fail fast is better than
// silently running with a broken config.
func Load() *Config {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		panic("DATABASE_URL environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET environment variable is required")
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	bcryptCost := 12
	if costStr := os.Getenv("BCRYPT_COST"); costStr != "" {
		cost, err := strconv.Atoi(costStr)
		if err != nil {
			panic(fmt.Sprintf("BCRYPT_COST must be an integer, got: %s", costStr))
		}
		bcryptCost = cost
	}

	return &Config{
		DatabaseURL: databaseURL,
		JWTSecret:   jwtSecret,
		ServerPort:  port,
		BcryptCost:  bcryptCost,
	}
}
