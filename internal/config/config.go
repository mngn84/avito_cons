package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func New() (*Config, error) {
	cfg := &Config{
		Webhook: WebhookConfig{
			Host: getEnv("WEBHOOK_HOST", "0:0:0:0"),
			Port: getEnv("WEBHOOK_PORT", "10000"),
		},
		OpenAI: OpenAIConfig{
			ApiKey:       getEnv("OPENAI_API_KEY", ""),
			Model:        getEnv("OPENAI_MODEL", "gpt-4o"),
			ApiUrl:       getEnv("OPENAI_URL", "https://api.openai.com/v1/"),
			SystemPrompt: getEnv("OPENAI_PROMPT", "You are a helpful assistant."),
			Temperature:  getFloat32("OPENAI_TEMPERATURE", 0.5),
			Timeout:      getDuration("OPENAI_TIMEOUT", 3*time.Second),
		},
		Avito: AvitoConfig{
			Token:   getEnv("AVITO_TOKEN", ""),
			ApiUrl:  getEnv("AVITO_API_URL", "https://api.avito.ru"),
			timeout: getDuration("AVITO_TIMEOUT", 3*time.Second),
		},
		DB: PgConfig{
			URL:      getEnv("POSTGRES_URL", ""),
			HistoryLimit: getInt("POSTGRES_LIMIT", 5),
			// Host:     getEnv("POSTGRES_HOST", "localhost"),
			// Port:     getEnv("POSTGRES_PORT", "5432"),
			// User:     getEnv("POSTGRES_USER", "postgres"),
			// Password: getEnv("POSTGRES_PASSWORD", ""),
			// DbName:   getEnv("POSTGRES_DB", "chatbot"),
			// SSLMode:  getEnv("POSTGRES_SSL_MODE", "disable"),
		},
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.OpenAI.ApiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY is required")
	}
	if c.Avito.Token == "" {
		return fmt.Errorf("AVITO_TOKEN is required")
	}
	return nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getFloat32(key string, defaultVal float32) float32 {
    if val := os.Getenv(key); val != "" {
        if f, err := strconv.ParseFloat(val, 32); err == nil {
            return float32(f)
        }
    }
    return defaultVal
}

func getDuration(key string, defaultVal time.Duration) time.Duration {
    if val := os.Getenv(key); val != "" {
        if d, err := time.ParseDuration(val); err == nil {
            return d
        }
    }
    return defaultVal
}

func getInt(key string, defaultVal int) int {
    if val := os.Getenv(key); val != "" {
        if i, err := strconv.Atoi(val); err == nil {
            return i
        }
    }
    return defaultVal
}