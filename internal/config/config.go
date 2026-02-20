package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv   string
	HTTPAddr string

	MySQLDSN string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	AdminToken       string
	CacheProductTTL  time.Duration
	CORSAllowOrigins []string
}

func MustLoad() Config {
	return Load()
}

func Load() Config {
	return Config{
		AppEnv:   getEnv("APP_ENV", "dev"),
		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),

		MySQLDSN: getEnv("MYSQL_DSN", "catalog-user:rootp@55word@tcp(192.168.68.160:3306)/catalog-db?parseTime=true&multiStatements=true"),

		RedisAddr:     getEnv("REDIS_ADDR", "192.168.68.160:6399"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		AdminToken:       getEnv("ADMIN_TOKEN", "dev-admin-token"),
		CacheProductTTL:  getEnvDuration("CACHE_PRODUCT_TTL", 2*time.Minute),
		CORSAllowOrigins: getEnvSlice("CORS_ALLOW_ORIGINS", "http://localhost:3000"),
	}
}

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func getEnvSlice(key, def string) []string {
	v := os.Getenv(key)
	if v == "" {
		v = def
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
