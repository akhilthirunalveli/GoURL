package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig
	DB     DBConfig
	Redis  RedisConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Addr     string
	Password string
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Replace dots with underscores in env variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No .env file found, relying on environment variables")
		} else {
			log.Printf("Error reading config file: %v", err)
		}
	}

	// Set defaults
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("ENV", "development")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_SSLMODE", "require")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")

	var cfg Config

	cfg.Server.Port = viper.GetString("SERVER_PORT")
	cfg.Server.Env = viper.GetString("ENV")

	cfg.DB.Host = viper.GetString("DB_HOST")
	cfg.DB.Port = viper.GetString("DB_PORT")
	cfg.DB.User = viper.GetString("DB_USER")
	cfg.DB.Password = viper.GetString("DB_PASSWORD")
	cfg.DB.Name = viper.GetString("DB_NAME")
	cfg.DB.SSLMode = viper.GetString("DB_SSLMODE")

	// Validate required DB configuration values
	if cfg.DB.User == "" {
		log.Fatalf("required configuration DB_USER is missing or empty")
	}
	if cfg.DB.Password == "" {
		log.Fatalf("required configuration DB_PASSWORD is missing or empty")
	}
	if cfg.DB.Name == "" {
		log.Fatalf("required configuration DB_NAME is missing or empty")
	}
	cfg.Redis.Addr = viper.GetString("REDIS_ADDR")
	cfg.Redis.Password = viper.GetString("REDIS_PASSWORD")

	return &cfg
}
