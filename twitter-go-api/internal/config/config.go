package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/spf13/viper"
)

type Config struct {
	Environment              string `mapstructure:"ENVIRONMENT"`
	DBSource                 string `mapstructure:"DATABASE_URL"`
	HTTPServerAddress        string `mapstructure:"HTTP_SERVER_ADDRESS"`
	DBMaxConns               int32  `mapstructure:"DB_MAX_CONNS"`
	DBMinConns               int32  `mapstructure:"DB_MIN_CONNS"`
	DBMaxConnLifetimeMinutes int    `mapstructure:"DB_MAX_CONN_LIFETIME_MINUTES"`
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
	S3BucketName             string `mapstructure:"S3_BUCKET_NAME"`
	S3Region                 string `mapstructure:"S3_REGION"`
	CloudFrontDomain         string `mapstructure:"CLOUDFRONT_DOMAIN"`
	GatewaySecret            string `mapstructure:"GATEWAY_SECRET"`
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
	viper.SetDefault("MAX_MEDIA_BYTES", int64(100<<20))  // 100 MiB
	viper.SetDefault("MAX_AVATAR_BYTES", int64(5<<20))   // 5 MiB
	// Explicitly bind environment variables so viper.Unmarshal detects them
	// without needing a physical app.env file (which is excluded in CI/CD).
	viper.BindEnv("ENVIRONMENT")
	viper.BindEnv("DATABASE_URL")
	viper.BindEnv("HTTP_SERVER_ADDRESS")
	viper.BindEnv("DB_MAX_CONNS")
	viper.BindEnv("DB_MIN_CONNS")
	viper.BindEnv("DB_MAX_CONN_LIFETIME_MINUTES")
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
	viper.BindEnv("S3_BUCKET_NAME")
	viper.BindEnv("S3_REGION")
	viper.BindEnv("CLOUDFRONT_DOMAIN")
	viper.BindEnv("GATEWAY_SECRET")
	viper.BindEnv("REDIS_ADDRESS")
	viper.BindEnv("REDIS_PASSWORD")

	viper.AutomaticEnv()

	env := os.Getenv("ENVIRONMENT")
	if env == "production" {
		loadFromSSM()
	}

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

func loadFromSSM() {
	prefix := "/chmtwt/prod/"
	
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Printf("Unable to load AWS config for SSM: %v\n", err)
		return
	}

	client := ssm.NewFromConfig(cfg)
	paginator := ssm.NewGetParametersByPathPaginator(client, &ssm.GetParametersByPathInput{
		Path:           &prefix,
		WithDecryption: func() *bool { b := true; return &b }(),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			fmt.Printf("Error fetching SSM parameters: %v\n", err)
			return
		}
		for _, p := range page.Parameters {
			key := strings.TrimPrefix(*p.Name, prefix)
			viper.Set(key, *p.Value)
		}
	}
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
	if strings.TrimSpace(c.GoogleClientID) == "" {
		return fmt.Errorf("GOOGLE_CLIENT_ID is required in production")
	}
	if len(strings.TrimSpace(c.TokenSymmetricKey)) < 32 {
		return fmt.Errorf("TOKEN_SYMMETRIC_KEY must be at least 32 characters in production")
	}
	if strings.TrimSpace(c.S3BucketName) == "" {
		return fmt.Errorf("S3_BUCKET_NAME is required in production")
	}
	if strings.TrimSpace(c.S3Region) == "" {
		return fmt.Errorf("S3_REGION is required in production")
	}
	if strings.TrimSpace(c.CloudFrontDomain) == "" {
		return fmt.Errorf("CLOUDFRONT_DOMAIN is required in production")
	}

	return nil
}
