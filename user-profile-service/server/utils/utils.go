package utils

import "os"

// GetEnv reads an environment variable with a fallback default.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
