package main

import (
	"os"
)

type Config struct {
	URL             string
	Username        string
	Password        string
	DeleteFiles     bool
	PollIntervalSec int
}

func LoadConfig() Config {
	cfg := Config{
		URL:             getEnv("QB_URL", "http://localhost:8080"),
		Username:        getEnv("QB_USER", "admin"),
		Password:        getEnv("QB_PASS", "adminadmin"),
		DeleteFiles:     getEnv("QB_DELETE_FILES", "true") == "true",
		PollIntervalSec: 30,
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
