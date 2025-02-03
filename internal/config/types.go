package config

import "time"

type Config struct {
 Webhook WebhookConfig
 OpenAI OpenAIConfig
 Avito AvitoConfig
 DB PgConfig
}

type WebhookConfig struct {
	Host string
	Port string
}

type OpenAIConfig struct {
	ApiKey string
	Model string
	ApiUrl string
	SystemPrompt string
	Temperature float32
	Timeout time.Duration
}

type AvitoConfig struct {
	Token string
	ApiUrl string
	timeout time.Duration
}

type PgConfig struct {
	URL string
	Host string
	Port string
	User string
	Password string
	DbName string
	SSLMode string
	HistoryLimit int
}