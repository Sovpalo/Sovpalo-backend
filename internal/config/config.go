package config

import (
	"os"
	"strconv"
	"strings"
)

// Config stores application settings loaded from environment variables.
type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	LLMProvider   string
	LLMTimeoutSec int

	YandexAPIKey   string
	YandexFolderID string
	YandexModel    string

	APNSKeyID          string
	APNSTeamID         string
	APNSBundleID       string
	APNSPrivateKeyPath string
	APNSPrivateKey     string
	APNSProduction     bool
}

func Load() Config {
	loadDotEnv(".env")
	return Config{
		Port:       getEnv("APP_PORT", "8000"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "sovpalo"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		LLMProvider:   getEnv("LLM_PROVIDER", ""),
		LLMTimeoutSec: getEnvInt("LLM_TIMEOUT_SEC", 15),

		YandexAPIKey:   getEnv("YANDEX_API_KEY", ""),
		YandexFolderID: getEnv("YANDEX_FOLDER_ID", ""),
		YandexModel:    getEnv("YANDEX_MODEL", "yandexgpt/latest"),

		APNSKeyID:          getEnv("APNS_KEY_ID", ""),
		APNSTeamID:         getEnv("APNS_TEAM_ID", ""),
		APNSBundleID:       getEnv("APNS_BUNDLE_ID", ""),
		APNSPrivateKeyPath: getEnv("APNS_PRIVATE_KEY_PATH", ""),
		APNSPrivateKey:     getEnv("APNS_PRIVATE_KEY", ""),
		APNSProduction:     getEnvBool("APNS_PRODUCTION", false),
	}
}

func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, val)
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	return value == "true" || value == "1" || value == "yes"
}
