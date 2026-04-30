package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv               string
	AppPort              string
	AppURL               string
	FrontendURL          string
	DatabaseURL          string
	JWTAccessSecret      string
	JWTAccessTTL         time.Duration
	RefreshTokenTTL      time.Duration
	CookieDomain         string
	CookieSecure         bool
	BcryptCost           int
	DefaultOwnerEmail    string
	DefaultOwnerPassword string
	DefaultCompanyName   string
}

func Load() Config {
	_ = godotenv.Load()
	return Config{
		AppEnv:               env("APP_ENV", "local"),
		AppPort:              env("APP_PORT", "8080"),
		AppURL:               env("APP_URL", "http://localhost:8080"),
		FrontendURL:          env("FRONTEND_URL", "http://localhost:3000"),
		DatabaseURL:          mustEnv("DATABASE_URL"),
		JWTAccessSecret:      mustEnv("JWT_ACCESS_SECRET"),
		JWTAccessTTL:         time.Duration(envInt("JWT_ACCESS_TTL_MINUTES", 15)) * time.Minute,
		RefreshTokenTTL:      time.Duration(envInt("REFRESH_TOKEN_TTL_DAYS", 30)) * 24 * time.Hour,
		CookieDomain:         env("COOKIE_DOMAIN", "localhost"),
		CookieSecure:         envBool("COOKIE_SECURE", false),
		BcryptCost:           envInt("BCRYPT_COST", 12),
		DefaultOwnerEmail:    env("DEFAULT_OWNER_EMAIL", "owner@example.com"),
		DefaultOwnerPassword: env("DEFAULT_OWNER_PASSWORD", "ChangeMe123!"),
		DefaultCompanyName:   env("DEFAULT_COMPANY_NAME", "Demo Company"),
	}
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic("missing env: " + k)
	}
	return v
}
func envInt(k string, d int) int {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return d
	}
	return i
}
func envBool(k string, d bool) bool {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return d
	}
	return b
}
