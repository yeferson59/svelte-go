package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Env struct {
	Port        string
	DatabaseURL string
}

func (c *Config) LoadEnvs() *Env {
	_ = godotenv.Load()

	return &Env{
		Port:        c.GetString("PORT", "8080"),
		DatabaseURL: c.GetString("DATABASE_URL", ""),
	}
}

func (Config) GetString(key, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(strings.ToUpper(key)))
	if value != "" {
		return value
	}

	return defaultValue
}

func (Config) GetInt(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(strings.ToUpper(key)))
	if value == "" {
		return defaultValue
	}

	IntValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return IntValue
}

func (Config) GetInt64(key string, defaultValue int64) int64 {
	value := strings.TrimSpace(os.Getenv(strings.ToUpper(key)))
	if value == "" {
		return defaultValue
	}

	int64Value, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}

	return int64Value
}

func (Config) GetBool(key string, defaultValue bool) bool {
	value := strings.TrimSpace(os.Getenv(strings.ToUpper(key)))
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolValue
}

func (Config) GetFloat64(key string, defaultValue float64) float64 {
	value := strings.TrimSpace(os.Getenv(strings.ToUpper(key)))
	if value == "" {
		return defaultValue
	}

	float64Value, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}

	return float64Value
}

func (Config) GetSlice(key string, defaultValue ...string) []string {
	value := strings.TrimSpace(os.Getenv(strings.ToUpper(key)))
	if value == "" {
		return defaultValue
	}

	keySplit := ","

	if !strings.Contains(value, keySplit) {
		return []string{value}
	}

	return strings.Split(value, keySplit)
}
