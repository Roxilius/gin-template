package config

import (
	"strings"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HTTP.Port != "8080" || cfg.App.Env != EnvDevelopment || cfg.HTTP.MaxBodyBytes != 1<<20 {
		t.Errorf("unexpected defaults: %+v", cfg.HTTP)
	}
}

func TestLoad_ParsesValues(t *testing.T) {
	t.Setenv("APP_PORT", "9000")
	t.Setenv("SHUTDOWN_TIMEOUT", "3s")
	t.Setenv("SWAGGER_ENABLED", "true")
	t.Setenv("CORS_ALLOWED_ORIGINS", " http://a.com, ,http://b.com ")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HTTP.Port != "9000" || cfg.App.ShutdownTimeout.Seconds() != 3 || !cfg.Swagger.Enabled {
		t.Errorf("values not parsed: %+v", cfg)
	}
	if got := strings.Join(cfg.HTTP.CORSAllowedOrigins, "|"); got != "http://a.com|http://b.com" {
		t.Errorf("origins = %q", got)
	}
}

func TestLoad_Validation(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{"invalid env", map[string]string{"APP_ENV": "staging"}, "APP_ENV must be one of"},
		{"prod without db password", map[string]string{"APP_ENV": EnvProduction, "DB_SSLMODE": "require"}, "DB_PASSWORD is required"},
		{"prod without ssl", map[string]string{"APP_ENV": EnvProduction, "DB_PASSWORD": "x", "DB_SSLMODE": "disable"}, "DB_SSLMODE=disable"},
		{"invalid body limit", map[string]string{"HTTP_MAX_BODY_BYTES": "0"}, "HTTP_MAX_BODY_BYTES"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err = %v, want containing %q", err, tc.wantErr)
			}
		})
	}
}

func TestDSN_EscapesPassword(t *testing.T) {
	d := DBConfig{Host: "localhost", Port: "5432", User: "app", Password: "p@ss w/rd", Name: "app", SSLMode: "disable"}
	want := "postgres://app:p%40ss%20w%2Frd@localhost:5432/app?sslmode=disable"
	if got := d.DSN(); got != want {
		t.Errorf("DSN = %s, want %s", got, want)
	}
	if got := d.MigrateURL(); !strings.HasPrefix(got, "pgx5://") {
		t.Errorf("MigrateURL = %s", got)
	}
}
