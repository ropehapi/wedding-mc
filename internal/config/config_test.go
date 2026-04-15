package config

import (
	"strings"
	"testing"
	"time"
)

// clearEnv zera todas as env vars usadas por Load() pro teste começar num estado limpo.
// Usa t.Setenv para que o cleanup seja automático ao final do teste.
// Nota: Load() chama godotenv.Load() que lê .env do cwd. Como os testes rodam
// de internal/config/, não existe .env lá, então não há contaminação.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"DATABASE_URL", "JWT_SECRET", "JWT_EXPIRY", "REFRESH_EXPIRY",
		"STORAGE_DRIVER", "LOCAL_STORAGE_PATH", "S3_BUCKET", "S3_REGION",
		"PORT", "ALLOWED_ORIGINS",
	} {
		t.Setenv(k, "")
	}
}

func TestLoad_Defaults(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/db")
	t.Setenv("JWT_SECRET", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port: got %q, want 8080", cfg.Port)
	}
	if cfg.JWTExpiry != time.Hour {
		t.Errorf("JWTExpiry: got %v, want 1h", cfg.JWTExpiry)
	}
	if cfg.RefreshExpiry != 168*time.Hour {
		t.Errorf("RefreshExpiry: got %v, want 168h", cfg.RefreshExpiry)
	}
	if cfg.StorageDriver != "local" {
		t.Errorf("StorageDriver: got %q, want local", cfg.StorageDriver)
	}
	if cfg.LocalStoragePath != "./uploads" {
		t.Errorf("LocalStoragePath: got %q, want ./uploads", cfg.LocalStoragePath)
	}
	if cfg.AllowedOrigins != "*" {
		t.Errorf("AllowedOrigins: got %q, want *", cfg.AllowedOrigins)
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	clearEnv(t)
	t.Setenv("JWT_SECRET", "secret")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected DATABASE_URL error, got %v", err)
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/db")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Fatalf("expected JWT_SECRET error, got %v", err)
	}
}

func TestLoad_InvalidJWTExpiry(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/db")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("JWT_EXPIRY", "not-a-duration")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "JWT_EXPIRY") {
		t.Fatalf("expected JWT_EXPIRY error, got %v", err)
	}
}

func TestLoad_InvalidRefreshExpiry(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/db")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("REFRESH_EXPIRY", "xxx")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "REFRESH_EXPIRY") {
		t.Fatalf("expected REFRESH_EXPIRY error, got %v", err)
	}
}

func TestLoad_InvalidStorageDriver(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/db")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("STORAGE_DRIVER", "ftp")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "STORAGE_DRIVER") {
		t.Fatalf("expected STORAGE_DRIVER error, got %v", err)
	}
}

func TestLoad_S3RequiresBucketAndRegion(t *testing.T) {
	t.Run("missing bucket", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("DATABASE_URL", "postgres://localhost/db")
		t.Setenv("JWT_SECRET", "secret")
		t.Setenv("STORAGE_DRIVER", "s3")
		t.Setenv("S3_REGION", "us-east-1")

		_, err := Load()
		if err == nil || !strings.Contains(err.Error(), "S3_BUCKET") {
			t.Fatalf("expected S3_BUCKET error, got %v", err)
		}
	})

	t.Run("missing region", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("DATABASE_URL", "postgres://localhost/db")
		t.Setenv("JWT_SECRET", "secret")
		t.Setenv("STORAGE_DRIVER", "s3")
		t.Setenv("S3_BUCKET", "my-bucket")

		_, err := Load()
		if err == nil || !strings.Contains(err.Error(), "S3_REGION") {
			t.Fatalf("expected S3_REGION error, got %v", err)
		}
	})

	t.Run("both present", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("DATABASE_URL", "postgres://localhost/db")
		t.Setenv("JWT_SECRET", "secret")
		t.Setenv("STORAGE_DRIVER", "s3")
		t.Setenv("S3_BUCKET", "my-bucket")
		t.Setenv("S3_REGION", "us-east-1")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.S3Bucket != "my-bucket" || cfg.S3Region != "us-east-1" {
			t.Errorf("s3 fields not set: %+v", cfg)
		}
	})
}

func TestLoad_CustomValues(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/db")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("JWT_EXPIRY", "30m")
	t.Setenv("REFRESH_EXPIRY", "72h")
	t.Setenv("PORT", "9090")
	t.Setenv("ALLOWED_ORIGINS", "https://example.com")
	t.Setenv("LOCAL_STORAGE_PATH", "/tmp/uploads")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.JWTExpiry != 30*time.Minute {
		t.Errorf("JWTExpiry: got %v, want 30m", cfg.JWTExpiry)
	}
	if cfg.RefreshExpiry != 72*time.Hour {
		t.Errorf("RefreshExpiry: got %v, want 72h", cfg.RefreshExpiry)
	}
	if cfg.Port != "9090" {
		t.Errorf("Port: got %q, want 9090", cfg.Port)
	}
	if cfg.AllowedOrigins != "https://example.com" {
		t.Errorf("AllowedOrigins: got %q", cfg.AllowedOrigins)
	}
	if cfg.LocalStoragePath != "/tmp/uploads" {
		t.Errorf("LocalStoragePath: got %q", cfg.LocalStoragePath)
	}
}
