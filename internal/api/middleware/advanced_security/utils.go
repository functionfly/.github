package advanced_security

import (
	"os"
	"strconv"
	"strings"
)

// Environment variable helper functions
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvStringSlice(key, defaultValue string) []string {
	if value := os.Getenv(key); value != "" {
		if value != "" {
			return strings.Split(value, ",")
		}
	}
	if defaultValue != "" {
		return strings.Split(defaultValue, ",")
	}
	return []string{}
}