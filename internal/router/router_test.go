package router

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	_ "github.com/Roxilius/gin-template/docs" // registrasi spesifikasi Swagger untuk test doc.json
	"github.com/Roxilius/gin-template/internal/config"
	"github.com/Roxilius/gin-template/internal/handler"
)

func newTestRouter(t *testing.T, pingErr error, mutate ...func(*config.Config)) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		App:  config.AppConfig{Name: "test-app", Env: config.EnvTest},
		HTTP: config.HTTPConfig{MaxBodyBytes: 1 << 20},
	}
	for _, m := range mutate {
		m(&cfg)
	}

	ping := func(context.Context) error { return pingErr }
	r, err := New(Deps{
		Config: cfg,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Health: handler.NewHealthHandler(cfg.App.Name, ping),
	})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func do(r http.Handler, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHealthEndpoints(t *testing.T) {
	r := newTestRouter(t, nil)

	for _, path := range []string{"/health", "/ready", "/api/v1/ping"} {
		if w := do(r, http.MethodGet, path, nil); w.Code != http.StatusOK {
			t.Errorf("GET %s: status %d, body %s", path, w.Code, w.Body)
		}
	}

	w := do(r, http.MethodGet, "/api/v1/ping", nil)
	var body struct {
		Success bool              `json:"success"`
		Data    map[string]string `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || !body.Success || body.Data["message"] != "pong" {
		t.Errorf("ping body = %s", w.Body)
	}
}

func TestReady_DatabaseDown(t *testing.T) {
	r := newTestRouter(t, errors.New("dial tcp 10.0.0.5:5432: connection refused"))

	w := do(r, http.MethodGet, "/ready", nil)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d, want 503", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"database":"down"`) {
		t.Errorf("body = %s", w.Body)
	}
	if strings.Contains(w.Body.String(), "10.0.0.5") {
		t.Errorf("internal error leaked: %s", w.Body)
	}
	// Liveness tidak boleh ikut gagal hanya karena database mati.
	if w := do(r, http.MethodGet, "/health", nil); w.Code != http.StatusOK {
		t.Errorf("/health status %d, want 200", w.Code)
	}
}

func TestNotFoundAndMethodNotAllowed(t *testing.T) {
	r := newTestRouter(t, nil)

	if w := do(r, http.MethodGet, "/nope", nil); w.Code != http.StatusNotFound || !strings.Contains(w.Body.String(), "route not found") {
		t.Errorf("404: status %d, body %s", w.Code, w.Body)
	}
	if w := do(r, http.MethodPost, "/health", nil); w.Code != http.StatusMethodNotAllowed {
		t.Errorf("405: status %d", w.Code)
	}
}

func TestMiddleware_RequestIDAndSecurityHeaders(t *testing.T) {
	r := newTestRouter(t, nil)

	w := do(r, http.MethodGet, "/health", nil)
	if len(w.Header().Get("X-Request-ID")) != 32 {
		t.Errorf("generated request id = %q", w.Header().Get("X-Request-ID"))
	}
	if w.Header().Get("X-Content-Type-Options") != "nosniff" || w.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("security headers missing: %v", w.Header())
	}

	w = do(r, http.MethodGet, "/health", map[string]string{"X-Request-ID": "abc-123"})
	if got := w.Header().Get("X-Request-ID"); got != "abc-123" {
		t.Errorf("valid client request id not kept: %q", got)
	}

	w = do(r, http.MethodGet, "/health", map[string]string{"X-Request-ID": "evil\" injected=1"})
	if got := w.Header().Get("X-Request-ID"); got == "evil\" injected=1" || len(got) != 32 {
		t.Errorf("unsafe client request id accepted: %q", got)
	}
}

func TestCORS(t *testing.T) {
	r := newTestRouter(t, nil, func(c *config.Config) {
		c.HTTP.CORSAllowedOrigins = []string{"http://localhost:3000"}
	})

	w := do(r, http.MethodGet, "/health", map[string]string{"Origin": "http://localhost:3000"})
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("allowed origin: header = %q", got)
	}
	w = do(r, http.MethodGet, "/health", map[string]string{"Origin": "http://evil.com"})
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("disallowed origin got header %q", got)
	}
}

func TestSwaggerToggle(t *testing.T) {
	if w := do(newTestRouter(t, nil), http.MethodGet, "/swagger/index.html", nil); w.Code != http.StatusNotFound {
		t.Errorf("swagger disabled: status %d, want 404", w.Code)
	}
	on := newTestRouter(t, nil, func(c *config.Config) { c.Swagger.Enabled = true })
	if w := do(on, http.MethodGet, "/swagger/index.html", nil); w.Code != http.StatusOK {
		t.Errorf("swagger enabled: status %d, want 200", w.Code)
	}
	if w := do(on, http.MethodGet, "/swagger/doc.json", nil); w.Code != http.StatusOK {
		t.Errorf("swagger doc.json: status %d, want 200", w.Code)
	}
	if w := do(on, http.MethodGet, "/swagger/", nil); w.Code != http.StatusFound || w.Header().Get("Location") != "/swagger/index.html" {
		t.Errorf("GET /swagger/: status %d location %q, want 302 -> /swagger/index.html", w.Code, w.Header().Get("Location"))
	}
	if w := do(on, http.MethodGet, "/swagger", nil); w.Code != http.StatusMovedPermanently || w.Header().Get("Location") != "/swagger/" {
		t.Errorf("GET /swagger: status %d location %q, want 301 -> /swagger/", w.Code, w.Header().Get("Location"))
	}
}
