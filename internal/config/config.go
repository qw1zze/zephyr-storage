package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort     string
	IPFSApiURL     string
	IPFSGatewayURL string
	MaxBlobSizeMB  int64
}

func Load() *Config {
	maxBlob, err := strconv.ParseInt(getEnv("MAX_BLOB_SIZE_MB", "10"), 10, 64)
	if err != nil {
		maxBlob = 10
	}

	return &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		IPFSApiURL:     getEnv("IPFS_API_URL", "http://localhost:5001"),
		IPFSGatewayURL: getEnv("IPFS_GATEWAY_URL", "http://localhost:8080"),
		MaxBlobSizeMB:  maxBlob,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
