package pgsql

import (
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL    string
	Port     int
	DBName   string
	User     string
	Password string
}

func NewConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("DB: .env file not loaded (%v), using environment variables / defaults", err)
	}

	return Config{
		DBURL:    getEnvWithFallback([]string{"DB_HOST", "DB_URL"}, "127.0.0.1"),
		Port:     getEnvPort([]string{"DB_PORT", "PORT"}, 5432),
		DBName:   getEnvWithFallback([]string{"DB_NAME"}, "tictactoe"),
		User:     getEnvWithFallback([]string{"DB_USER"}, "postgres"),
		Password: getEnvWithFallback([]string{"DB_PASSWORD", "PASSWORD"}, "postgres"),
	}
}

func (c Config) FormConnString() string {
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     net.JoinHostPort(c.DBURL, strconv.Itoa(c.Port)),
		Path:     c.DBName,
		RawQuery: "sslmode=disable",
	}
	return u.String()
}

func getEnvWithFallback(keys []string, defaultVal string) string {
	for _, key := range keys {
		if val := strings.TrimSpace(os.Getenv(key)); val != "" {
			log.Printf("DB: Argument %s loaded", key)
			return val
		}
	}
	log.Printf("DB: Argument %v not set. Falling back to default value %q", keys, defaultVal)
	return defaultVal
}

func getEnvPort(keys []string, defaultPort int) int {
	for _, key := range keys {
		raw := strings.TrimSpace(os.Getenv(key))
		if raw == "" {
			continue
		}
		port, err := strconv.Atoi(raw)
		if err != nil || port <= 0 || port > 65535 {
			log.Printf("DB: Invalid port in %s=%q, falling back to default %d", key, raw, defaultPort)
			return defaultPort
		}
		log.Printf("DB: Argument %s loaded (%d)", key, port)
		return port
	}
	log.Printf("DB: Port %v not set. Falling back to default %d", keys, defaultPort)
	return defaultPort
}
