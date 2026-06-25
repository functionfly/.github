package config

import "os"

func IsVectorSearchEnabled() bool {
	return os.Getenv("ENABLE_VECTOR_SEARCH") == "true"
}
