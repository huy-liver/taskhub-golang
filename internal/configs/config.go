package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for our application
type Config struct {
	AppEnv  string `mapstructure:"APP_ENV"`
	AppPort string `mapstructure:"APP_PORT"`
	AppName string `mapstructure:"APP_NAME"`

	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBSSLMode  string `mapstructure:"DB_SSLMODE"`

	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPort     string `mapstructure:"REDIS_PORT"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`

	JWTSecret           string `mapstructure:"JWT_SECRET"`
	JWTExpiresIn        string `mapstructure:"JWT_EXPIRES_IN"`
	JWTRefreshSecret    string `mapstructure:"JWT_REFRESH_SECRET"`
	JWTRefreshExpiresIn string `mapstructure:"JWT_REFRESH_EXPIRES_IN"`

	CORSAllowedOrigins string `mapstructure:"CORS_ALLOWED_ORIGINS"`

	RateLimitRequestsPerMinute     int `mapstructure:"RATE_LIMIT_REQUESTS_PER_MINUTE"`
	RateLimitAuthRequestsPerMinute int `mapstructure:"RATE_LIMIT_AUTH_REQUESTS_PER_MINUTE"`

	CacheTTLMinutes int `mapstructure:"CACHE_TTL_MINUTES"`
}

// LoadConfig loads configuration from .env file or environment variables
func LoadConfig() (*Config, error) {
	v := viper.New()

	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("APP_NAME", "TaskHub API")
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "5432")
	v.SetDefault("DB_USER", "postgres")
	v.SetDefault("DB_PASSWORD", "postgres")
	v.SetDefault("DB_NAME", "taskhub_db")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", "6379")
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("JWT_SECRET", "super_secret_jwt_access_key_change_in_production")
	v.SetDefault("JWT_EXPIRES_IN", "24h")
	v.SetDefault("JWT_REFRESH_SECRET", "super_secret_jwt_refresh_key_change_in_production")
	v.SetDefault("JWT_REFRESH_EXPIRES_IN", "168h")
	v.SetDefault("CORS_ALLOWED_ORIGINS", "*")
	v.SetDefault("RATE_LIMIT_REQUESTS_PER_MINUTE", 100)
	v.SetDefault("RATE_LIMIT_AUTH_REQUESTS_PER_MINUTE", 10)
	v.SetDefault("CACHE_TTL_MINUTES", 15)

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		// It's acceptable if .env is missing when environment variables are set directly
		fmt.Printf("Notice: .env file not found or could not be loaded, using environment/defaults: %v\n", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return &cfg, nil
}
