package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Environment              string `mapstructure:"ENVIRONMENT"`
	DBSource                 string `mapstructure:"DATABASE_URL"`
	HTTPServerAddress        string `mapstructure:"HTTP_SERVER_ADDRESS"`
	DBMaxConns               int32  `mapstructure:"DB_MAX_CONNS"`
	DBMinConns               int32  `mapstructure:"DB_MIN_CONNS"`
	DBMaxConnLifetimeMinutes int    `mapstructure:"DB_MAX_CONN_LIFETIME_MINUTES"`
	MaxMultipartMemoryBytes  int64  `mapstructure:"MAX_MULTIPART_MEMORY_BYTES"`
	MaxMediaBytes            int64  `mapstructure:"MAX_MEDIA_BYTES"`
	MaxAvatarBytes           int64  `mapstructure:"MAX_AVATAR_BYTES"`
	FrontendURL              string `mapstructure:"FRONTEND_URL"`
	CookieDomain             string `mapstructure:"COOKIE_DOMAIN"`
	CookieSameSite           string `mapstructure:"COOKIE_SAME_SITE"`
	CookieSecure             bool   `mapstructure:"COOKIE_SECURE"`
	TokenSymmetricKey        string `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	TokenDurationMinutes     int    `mapstructure:"TOKEN_DURATION_MINUTES"`
	RefreshTokenDurationDays int    `mapstructure:"REFRESH_TOKEN_DURATION_DAYS"`
	GoogleClientID           string `mapstructure:"GOOGLE_CLIENT_ID"`
	AzureStorageConnString   string `mapstructure:"AZURE_STORAGE_CONNECTION_STRING"`
	AzureStorageContainer    string `mapstructure:"AZURE_STORAGE_CONTAINER_NAME"`
	RedisAddress             string `mapstructure:"REDIS_ADDRESS"`
	RedisPassword            string `mapstructure:"REDIS_PASSWORD"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.SetDefault("FRONTEND_URL", "http://localhost:3000")
	viper.SetDefault("COOKIE_DOMAIN", "")
	viper.SetDefault("COOKIE_SAME_SITE", "Lax")
	viper.SetDefault("COOKIE_SECURE", false)
	viper.SetDefault("DB_MAX_CONNS", int32(25))
	viper.SetDefault("DB_MIN_CONNS", int32(0))
	viper.SetDefault("DB_MAX_CONN_LIFETIME_MINUTES", 5)
	viper.SetDefault("MAX_MULTIPART_MEMORY_BYTES", int64(32<<20)) // 32 MiB
	viper.SetDefault("MAX_MEDIA_BYTES", int64(100<<20))           // 100 MiB
	viper.SetDefault("MAX_AVATAR_BYTES", int64(5<<20))            // 5 MiB

	// Explicitly bind environment variables so viper.Unmarshal detects them
	// without needing a physical app.env file (which is excluded in CI/CD).
	viper.BindEnv("ENVIRONMENT")
	viper.BindEnv("DATABASE_URL")
	viper.BindEnv("HTTP_SERVER_ADDRESS")
	viper.BindEnv("DB_MAX_CONNS")
	viper.BindEnv("DB_MIN_CONNS")
	viper.BindEnv("DB_MAX_CONN_LIFETIME_MINUTES")
	viper.BindEnv("MAX_MULTIPART_MEMORY_BYTES")
	viper.BindEnv("MAX_MEDIA_BYTES")
	viper.BindEnv("MAX_AVATAR_BYTES")
	viper.BindEnv("FRONTEND_URL")
	viper.BindEnv("COOKIE_DOMAIN")
	viper.BindEnv("COOKIE_SAME_SITE")
	viper.BindEnv("COOKIE_SECURE")
	viper.BindEnv("TOKEN_SYMMETRIC_KEY")
	viper.BindEnv("TOKEN_DURATION_MINUTES")
	viper.BindEnv("REFRESH_TOKEN_DURATION_DAYS")
	viper.BindEnv("GOOGLE_CLIENT_ID")
	viper.BindEnv("AZURE_STORAGE_CONNECTION_STRING")
	viper.BindEnv("AZURE_STORAGE_CONTAINER_NAME")
	viper.BindEnv("REDIS_ADDRESS")
	viper.BindEnv("REDIS_PASSWORD")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		// It's ok if app.env doesn't exist, we will use auto env
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return
		}
	}

	err = viper.Unmarshal(&config)
	return
}

func (c Config) ValidateForRuntime() error {
	if strings.EqualFold(strings.TrimSpace(c.CookieSameSite), "none") && !c.CookieSecure {
		return fmt.Errorf("COOKIE_SECURE must be true when COOKIE_SAME_SITE=None")
	}

	if !strings.EqualFold(strings.TrimSpace(c.Environment), "production") {
		return nil
	}

	if strings.TrimSpace(c.FrontendURL) == "" {
		return fmt.Errorf("FRONTEND_URL is required in production")
	}
	if !c.CookieSecure {
		return fmt.Errorf("COOKIE_SECURE must be true in production")
	}
	if strings.Contains(strings.ToLower(c.DBSource), "sslmode=disable") {
		return fmt.Errorf("DATABASE_URL must use sslmode in production")
	}
	if strings.TrimSpace(c.AzureStorageConnString) == "" {
		return fmt.Errorf("AZURE_STORAGE_CONNECTION_STRING is required in production")
	}
	if strings.TrimSpace(c.AzureStorageContainer) == "" {
		return fmt.Errorf("AZURE_STORAGE_CONTAINER_NAME is required in production")
	}
	if strings.TrimSpace(c.GoogleClientID) == "" {
		return fmt.Errorf("GOOGLE_CLIENT_ID is required in production")
	}

	return nil
}
