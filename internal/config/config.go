package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DetectorServiceURL string
	UserServiceURL     string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("[config] no .env file found, reading from environment")
	}

	return Config{
		Port:               getEnv("PORT", "8080"),
		DetectorServiceURL: getEnv("DETECTOR_SERVICE_URL", ""),
		UserServiceURL:     getEnv("USER_SERVICE_URL", "http://localhost:8082"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
