package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port             int
	Env              string
	DatabaseURL      string
	ElasticsearchURL string
	OpenAIKey        string
	TEDBaseURL       string
	DadosBaseURL     string
	LogLevel         string
}

func Load() Config {
	return Config{
		Port:             getEnvInt("PORT", 8080),
		Env:              getEnvStr("ENV", "dev"),
		DatabaseURL:      getEnvStr("DATABASE_URL", "postgres://spotgov:spotgov_pass@localhost:5432/spotgov?sslmode=disable"),
		ElasticsearchURL: getEnvStr("ELASTICSEARCH_URL", "http://localhost:9200"),
		OpenAIKey:        getEnvStr("OPENAI_API_KEY", ""),
		TEDBaseURL:       getEnvStr("TED_BASE_URL", "https://api.ted.europa.eu"),
		DadosBaseURL:     getEnvStr("DADOS_BASE_URL", "https://dados.gov.pt/api/1"),
		LogLevel:         getEnvStr("LOG_LEVEL", "info"),
	}
}

func (c Config) IsDev() bool {
	return c.Env == "dev" || c.Env == "local"
}

func getEnvStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}
