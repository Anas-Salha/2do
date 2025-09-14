package config

import (
	"log"
	"os"
)

type Config struct {
	DBName     string
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	Port       string
}

func Load() Config {
	return Config{
		DBName:     getEnv("DB_NAME"),
		DBUser:     getEnv("DB_USER"),
		DBPassword: getEnv("DB_PASS"),
		DBHost:     getEnv("DB_HOST"),
		DBPort:     getEnv("DB_PORT"),
		Port:       getEnvDefault("PORT", "8080"),
	}
}

func getEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("Environment variable %s not set.", k)
	}

	return v
}

func getEnvDefault(k string, d string) string {
	v := os.Getenv(k)
	if v == "" {
		v = d
	}

	return v
}
