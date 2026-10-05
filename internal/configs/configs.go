package configs

import "os"

type Config struct {
	Port     string
	LogLevel string
	DB       DB
}

type DB struct {
	Host, Port, User, Password, Name string
}

func Load() Config {
	return Config{
		Port:     get("APP_PORT", "8083"),
		LogLevel: get("LOG_LEVEL", "info"),
		DB: DB{
			Host:     get("DB_HOST", "localhost"),
			Port:     get("DB_PORT", "5432"),
			User:     get("DB_USER", "postgres"),
			Password: get("DB_PASSWORD", ""),
			Name:     get("DB_NAME", "ab_system"),
		},
	}
}
func get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
