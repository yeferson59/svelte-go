package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Env struct {
	Enviroment    string
	Port          string
	PathMigration string
	DatabaseURL   string
	JWTSecret     string
	JWTDuration   time.Duration
	CORSEnabled   bool
	CORSOrigin    []string
}

func (c *Config) LoadEnvs() *Env {
	_ = godotenv.Load()

	return &Env{
		Enviroment:    c.getString("ENVIROMENT", "development"),
		Port:          c.getString("PORT", "8080"),
		PathMigration: c.getString("PATH_MIGRATION", "file://internal/migrations"),
		DatabaseURL:   c.getString("DATABASE_URL", ""),
		JWTSecret:     c.getString("JWT_SECRET", "secret"),
		JWTDuration:   c.getDuration("JWT_DURATION", time.Hour*24*7),
		CORSEnabled:   c.getBool("CORS_ENABLED", true),
		CORSOrigin:    c.getSlice("CORS_ORIGIN", "http://localhost:5173"),
	}
}

func (Config) getString(key, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(strings.ToUpper(key)))
	if value != "" {
		return value
	}

	return defaultValue
}

func (Config) getDuration(key string, defaultValue time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(strings.ToUpper(key)))
	if value == "" {
		return defaultValue
	}

	if strings.Contains(value, "s") {
		before, _, _ := strings.Cut(value, "s")
		int64Value, err := strconv.ParseInt(before, 10, 64)
		if err != nil {
			return defaultValue
		}

		return time.Second * time.Duration(int64Value)
	}

	if strings.Contains(value, "m") {
		before, _, _ := strings.Cut(value, "m")
		int64Value, err := strconv.ParseInt(before, 10, 64)
		if err != nil {
			return defaultValue
		}

		return time.Minute * time.Duration(int64Value)
	}

	if strings.Contains(value, "h") {
		before, _, _ := strings.Cut(value, "h")
		int64Value, err := strconv.ParseInt(before, 10, 64)
		if err != nil {
			return defaultValue
		}

		return time.Hour * time.Duration(int64Value)
	}

	before, _, found := strings.Cut(value, "d")
	if !found {
		return defaultValue
	}

	int64Value, err := strconv.ParseInt(before, 10, 64)
	if err != nil {
		return defaultValue
	}

	return time.Hour * 24 * time.Duration(int64Value)
}

/*
 * func (Config) getInt(key string, defaultValue int) int {
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
*/

func (Config) getInt64(key string, defaultValue int64) int64 {
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

func (Config) getBool(key string, defaultValue bool) bool {
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

/*
 * func (Config) getFloat64(key string, defaultValue float64) float64 {
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
*/

func (Config) getSlice(key string, defaultValue ...string) []string {
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
