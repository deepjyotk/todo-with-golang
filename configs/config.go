package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the entire application configuration.
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
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	DSN      string // Constructed manually later
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

// LoadConfig loads configuration from a .env file (if present) and environment variables.
func LoadConfig() (*Config, error) {
	// Tell Viper the file we want to read is ".env"
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	// We also add the current directory as a config path to look for the .env file.
	viper.AddConfigPath(".")

	// Read in the config file. If it's not found, we skip it (not an error).
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Enable reading from environment variables:
	viper.AutomaticEnv()
	// Replace dots with underscores so viper will look for e.g. SERVER_PORT
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// --- Bind server config ---
	_ = viper.BindEnv("server.port")

	// --- Bind database config ---
	_ = viper.BindEnv("database.host")
	_ = viper.BindEnv("database.port")
	_ = viper.BindEnv("database.user")
	_ = viper.BindEnv("database.password")
	_ = viper.BindEnv("database.name")

	// --- Bind JWT ---
	_ = viper.BindEnv("jwt.secret")

	// --- Bind S3 ---
	_ = viper.BindEnv("s3.bucket")
	_ = viper.BindEnv("s3.region")
	_ = viper.BindEnv("s3.access_key")
	_ = viper.BindEnv("s3.secret_key")

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	// Construct the DSN manually from the DB fields read above
	config.Database.DSN = fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
	)

	return &config, nil
}
