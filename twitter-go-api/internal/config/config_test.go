package config

import "testing"

func TestValidateForRuntime_SameSiteNoneRequiresSecure(t *testing.T) {
	t.Parallel()

	cfg := Config{CookieSameSite: "None", CookieSecure: false}
	if err := cfg.ValidateForRuntime(); err == nil {
		t.Fatal("expected error when SameSite=None without Secure")
	}
}

func TestValidateForRuntime_SameSiteNoneWithSecureOK(t *testing.T) {
	t.Parallel()

	cfg := Config{CookieSameSite: "None", CookieSecure: true}
	if err := cfg.ValidateForRuntime(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateForRuntime_DevEnvironmentNoOtherChecks(t *testing.T) {
	t.Parallel()

	cfg := Config{Environment: "development", CookieSameSite: "Lax"}
	if err := cfg.ValidateForRuntime(); err != nil {
		t.Fatalf("unexpected error in dev environment: %v", err)
	}
}

func TestValidateForRuntime_ProductionRequiresFrontendURL(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Environment:            "production",
		CookieSameSite:         "Lax",
		CookieSecure:           true,
		DBSource:          "postgresql://host/db?sslmode=require",
		S3BucketName:      "my-bucket",
		S3Region:          "ap-southeast-1",
		CloudFrontDomain:  "d123.cloudfront.net",
		GoogleClientID:    "client-id",
	}
	if err := cfg.ValidateForRuntime(); err == nil {
		t.Fatal("expected error when FRONTEND_URL not set in production")
	}
}

func TestValidateForRuntime_ProductionRequiresCookieSecure(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Environment:    "production",
		FrontendURL:    "https://example.com",
		CookieSameSite: "Lax",
		CookieSecure:   false,
	}
	if err := cfg.ValidateForRuntime(); err == nil {
		t.Fatal("expected error when COOKIE_SECURE=false in production")
	}
}

func TestValidateForRuntime_ProductionRejectsInsecureDB(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Environment:    "production",
		FrontendURL:    "https://example.com",
		CookieSameSite: "Lax",
		CookieSecure:   true,
		DBSource:       "postgresql://host/db?sslmode=disable",
	}
	if err := cfg.ValidateForRuntime(); err == nil {
		t.Fatal("expected error when DB uses sslmode=disable in production")
	}
}

func TestValidateForRuntime_ProductionAllValidPasses(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Environment:            "production",
		FrontendURL:            "https://example.com",
		CookieSameSite:         "Lax",
		CookieSecure:           true,
		DBSource:          "postgresql://host/db?sslmode=require",
		S3BucketName:      "my-bucket",
		S3Region:          "ap-southeast-1",
		CloudFrontDomain:  "d123.cloudfront.net",
		GoogleClientID:    "my-client-id.apps.googleusercontent.com",
		TokenSymmetricKey: "01234567890123456789012345678901",
	}
	if err := cfg.ValidateForRuntime(); err != nil {
		t.Fatalf("unexpected error for valid production config: %v", err)
	}
}
