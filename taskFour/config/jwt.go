package config

import "os"

var JWTSecret = []byte(getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"))

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
