package configs

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	S3       S3Config       `mapstructure:"s3"`
}

// ServerConfig holds the server-specific configuration.
type ServerConfig struct {
	Port string `mapstructure:"port"`
}

// DatabaseConfig holds the database configuration.
type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

// JWTConfig holds the JWT-specific configuration.
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

// S3Config holds the AWS S3 configuration.
type S3Config struct {
	Bucket    string `mapstructure:"bucket"`
	Region    string `mapstructure:"region"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
}

// LoadConfig loads configuration from the specified file path.
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	// Optionally, you can set the config type explicitly:
	// viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return &config, nil
}
