package config

import (
    "log"
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    AppPort       string
    LaravelAPIURL string
    DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    DBName     string
}

func LoadConfig() Config {
    err := godotenv.Load()

    if err != nil {
        log.Println("No .env file found, using system environment")
    }

    return Config{
        AppPort:       getEnv("APP_PORT", "8080"),
        LaravelAPIURL: getEnv("LARAVEL_API_URL", "http://127.0.0.1:8000"),
        DBHost:     getEnv("DB_HOST", "127.0.0.1"),
        DBPort:     getEnv("DB_PORT", "3306"),
        DBUser:     getEnv("DB_USER", "root"),
        DBPassword: getEnv("DB_PASSWORD", ""),
        DBName:     getEnv("DB_NAME", ""),
    }
}

func getEnv(key, fallback string) string {
    value := os.Getenv(key)

    if value == "" {
        return fallback
    }

    return value
}