// Package config memuat dan memvalidasi konfigurasi aplikasi dari environment
// variable (dan file .env saat development).
package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	EnvDevelopment = "development"
	EnvTest        = "test"
	EnvProduction  = "production"
)

type Config struct {
	App     AppConfig
	HTTP    HTTPConfig
	DB      DBConfig
	Log     LogConfig
	Swagger SwaggerConfig
}

type AppConfig struct {
	Name            string
	Env             string
	ShutdownTimeout time.Duration
}

type HTTPConfig struct {
	Port               string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	TrustedProxies     []string
	CORSAllowedOrigins []string
	MaxBodyBytes       int64
}

type DBConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type LogConfig struct {
	Level string
}

type SwaggerConfig struct {
	Enabled bool
}

// Load membaca konfigurasi. File .env bersifat opsional: kalau ada, isinya
// dimuat, tapi tidak menimpa environment variable yang sudah di-set.
//
// Menambah konfigurasi baru (mis. Redis, JWT): tambahkan struct-nya di atas,
// isi di sini, dan validasi di validate().
func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		App: AppConfig{
			Name:            getString("APP_NAME", "gin-template"),
			Env:             getString("APP_ENV", EnvDevelopment),
			ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		HTTP: HTTPConfig{
			Port:               getString("APP_PORT", "8080"),
			ReadTimeout:        getDuration("HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:       getDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
			TrustedProxies:     getList("TRUSTED_PROXIES"),
			CORSAllowedOrigins: getList("CORS_ALLOWED_ORIGINS"),
			MaxBodyBytes:       int64(getInt("HTTP_MAX_BODY_BYTES", 1<<20)), // 1 MB
		},
		DB: DBConfig{
			Host:            getString("DB_HOST", "localhost"),
			Port:            getString("DB_PORT", "5432"),
			User:            getString("DB_USER", "app"),
			Password:        getString("DB_PASSWORD", ""),
			Name:            getString("DB_NAME", "app"),
			SSLMode:         getString("DB_SSLMODE", "disable"),
			MaxOpenConns:    getInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
		},
		Log: LogConfig{
			Level: getString("LOG_LEVEL", "info"),
		},
		Swagger: SwaggerConfig{
			Enabled: getBool("SWAGGER_ENABLED", false),
		},
	}

	return cfg, cfg.validate()
}

func (c Config) IsProduction() bool { return c.App.Env == EnvProduction }

func (c Config) validate() error {
	var errs []error

	switch c.App.Env {
	case EnvDevelopment, EnvTest, EnvProduction:
	default:
		errs = append(errs, fmt.Errorf("APP_ENV must be one of development|test|production, got %q", c.App.Env))
	}
	if c.IsProduction() && c.DB.Password == "" {
		errs = append(errs, errors.New("DB_PASSWORD is required in production"))
	}
	if c.IsProduction() && c.DB.SSLMode == "disable" {
		errs = append(errs, errors.New("DB_SSLMODE=disable is not allowed in production"))
	}
	if c.HTTP.MaxBodyBytes <= 0 {
		errs = append(errs, errors.New("HTTP_MAX_BODY_BYTES must be positive"))
	}

	return errors.Join(errs...)
}

// DSN mengembalikan connection string PostgreSQL dalam format URL.
// Dibangun dengan net/url supaya password berisi karakter khusus tetap aman.
func (d DBConfig) DSN() string { return d.url("postgres") }

// MigrateURL dipakai golang-migrate (driver pgx v5 memakai scheme "pgx5").
func (d DBConfig) MigrateURL() string { return d.url("pgx5") }

func (d DBConfig) url(scheme string) string {
	u := url.URL{
		Scheme:   scheme,
		User:     url.UserPassword(d.User, d.Password),
		Host:     net.JoinHostPort(d.Host, d.Port),
		Path:     "/" + d.Name,
		RawQuery: url.Values{"sslmode": {d.SSLMode}}.Encode(),
	}
	return u.String()
}

func getString(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v, err := strconv.Atoi(getString(key, "")); err == nil {
		return v
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	if v, err := strconv.ParseBool(getString(key, "")); err == nil {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v, err := time.ParseDuration(getString(key, "")); err == nil {
		return v
	}
	return fallback
}

// getList membaca nilai dipisah koma, misalnya "a, b,c" -> [a b c].
func getList(key string) []string {
	raw := getString(key, "")
	if raw == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
