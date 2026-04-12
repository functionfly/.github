package admin

import "os"

// getEnvOrDefault returns the environment variable value if set, otherwise a default.
func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	switch key {
	case "VITE_GOOGLE_ANALYTICS_ID":
		return "G-XXXXXXXXXX"
	case "VITE_HOTJAR_SITE_ID":
		return "0000000"
	default:
		return defaultValue
	}
}
