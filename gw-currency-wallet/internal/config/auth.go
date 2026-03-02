package config

import (
	"os"
	"time"
)

type AuthConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func LoadAuth() *AuthConfig {
	return &AuthConfig{
		Secret:     must("JWT_SECRET"),
		AccessTTL:  mustDuration("JWT_ACCESS_TTL"),
		RefreshTTL: mustDuration("JWT_REFRESH_TTL"),
	}
}

func must(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("missing env: " + key)
	}
	return v
}

func mustDuration(key string) time.Duration {
	d, err := time.ParseDuration(must(key))
	if err != nil {
		panic("invalid duration: " + key)
	}
	return d
}
