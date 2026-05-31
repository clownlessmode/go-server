package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	Postgres   PostgresConfig
	Auth       AuthConfig
	Proxy      ProxyConfig
	Rocketbank RocketbankConfig
	Beeline    BeelineConfig
	SMS        SMSConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSL      string
}

type AuthConfig struct {
	JWTSecret string
}

type ProxyConfig struct {
	Host           string
	Address        string
	CertDir        string
	ApkDir         string
	RocketbankLogs bool
	BeelineLogs    bool
}

type RocketbankConfig struct {
	Timezone string
}

type BeelineConfig struct {
	FirstPageType string
}

type SMSConfig struct {
	Enabled     bool
	AgentAPIKey string
}

func Load() Config {
	loadDotEnv(".env")

	return Config{
		Postgres: PostgresConfig{
			Host:     env("POSTGRES_HOST", "localhost"),
			Port:     env("POSTGRES_PORT", "5432"),
			User:     env("POSTGRES_USER", "postgres"),
			Password: env("POSTGRES_PASSWORD", "postgres"),
			Database: env("POSTGRES_DB", "mitm"),
			SSL:      env("POSTGRES_SSL", "false"),
		},
		Auth: AuthConfig{
			JWTSecret: env("JWT_SECRET", "dev-secret-change-me"),
		},
		Proxy: ProxyConfig{
			Host:           env("MITM_PROXY_HOST", "rebellion.proxy"),
			Address:        env("MITM_PROXY_ADDRESS", ":8888"),
			CertDir:        env("MITM_PROXY_CERT_DIR", "data/proxy"),
			ApkDir:         env("MITM_PROXY_APK_DIR", "web/apks"),
			RocketbankLogs: envBool("ROCKETBANK_LOGS", false),
			BeelineLogs:    envBool("BEELINE_LOGS", false),
		},
		Rocketbank: RocketbankConfig{
			Timezone: env("ROCKETBANK_TIMEZONE", "+0700"),
		},
		Beeline: BeelineConfig{
			FirstPageType: env("FIRST_PAGE_TYPE", env("FIRSTPAGETYPE", "1")),
		},
		SMS: SMSConfig{
			Enabled:     envBool("SMS_ENABLED", false),
			AgentAPIKey: env("SMS_AGENT_API_KEY", ""),
		},
	}
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}

	return value == "true" || value == "1" || value == "yes" || value == "on"
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" || os.Getenv(key) != "" {
			continue
		}

		_ = os.Setenv(key, value)
	}
}
