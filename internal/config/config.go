package config

import (
	"os"
)

// Config holds runtime configuration for the service.
type Config struct {
	// Addr is the listen address for the HTTP server, e.g. ":8080".
	Addr string
	// DataDir is the base directory for persisted image data.
	DataDir string
	// MaxUploadBytes limits the maximum upload size accepted by the API.
	MaxUploadBytes int64
}

// LoadFromEnv loads configuration from environment variables with sensible defaults.
// IMGAPI_ADDR, IMGAPI_DATA_DIR, IMGAPI_MAX_UPLOAD_MB
func LoadFromEnv() Config {
	addr := getEnvDefault("IMGAPI_ADDR", ":8080")
	dataDir := getEnvDefault("IMGAPI_DATA_DIR", "./data/images")
	maxUploadMB := int64FromEnv("IMGAPI_MAX_UPLOAD_MB", 25) // 25 MB default
	return Config{
		Addr:           addr,
		DataDir:        dataDir,
		MaxUploadBytes: maxUploadMB * 1024 * 1024,
	}
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func int64FromEnv(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		// best-effort parse; on error use default
		var out int64
		var sign int64 = 1
		for i := 0; i < len(v); i++ {
			c := v[i]
			if i == 0 && c == '-' {
				sign = -1
				continue
			}
			if c < '0' || c > '9' {
				return def
			}
			out = out*10 + int64(c-'0')
		}
		if out == 0 {
			return def
		}
		return out * sign
	}
	return def
}
