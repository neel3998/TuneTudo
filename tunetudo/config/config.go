package config

import (
	"os"
	"strings"
)

type Config struct {
	Port               string
	DatabasePath       string
	JWTSecret          string
	MaxUploadSize      int64
	StoragePath        string
	AllowedAudioTypes  []string
	AllowedImageTypes  []string
	TLS_KEY_FILE   string
	TLS_CERT_FILE  string
}

func LoadConfig() *Config {
	return &Config{
		Port:              getEnv("PORT", "2701"),
		DatabasePath:      getEnv("DATABASE_PATH", "./tunetudo.db"),
		JWTSecret:         getEnv("JWT_SECRET", "sup3rdup3rs3cr3t"),
		TLS_KEY_FILE:    getEnv("TLS_KEY_FILE", "./certs/server.key"),
		TLS_CERT_FILE:   getEnv("TLS_CERT_FILE", "./certs/server.crt"),
		MaxUploadSize:     50 * 1024 * 1024, // 50MB
		StoragePath:       getEnv("STORAGE_PATH", "./storage"),
		AllowedAudioTypes: []string{".mp4", ".wav", ".mp3"},
		AllowedImageTypes: []string{".jpg", ".jpeg", ".png"},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) IsAllowedAudioType(filename string) bool {
	for _, ext := range c.AllowedAudioTypes {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			return true
		}
	}
	return false
}

func (c *Config) IsAllowedImageType(filename string) bool {
	for _, ext := range c.AllowedImageTypes {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			return true
		}
	}
	return false
}