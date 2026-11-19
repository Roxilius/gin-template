package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Roxilius/gin-template/internal/domain"
	"github.com/Roxilius/gin-template/pkg/response"
)

func init() {
	gin.SetMode(gin.TestMode)
	RegisterValidator()
}

func TestError_MapsDomainErrors(t *testing.T) {
	tests := []struct {
		err  error
		want int
	}{
		{domain.ErrNotFound, http.StatusNotFound},
		{fmt.Errorf("find product: %w", domain.ErrNotFound), http.StatusNotFound},
		{domain.ErrConflict, http.StatusConflict},
		{domain.ErrUnauthorized, http.StatusUnauthorized},
		{domain.ErrForbidden, http.StatusForbidden},
		{errors.New("pq: connection refused at 10.0.0.5"), http.StatusInternalServerError},
	}
	for _, tc := range tests {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		Error(c, tc.err)

		if w.Code != tc.want {
			t.Errorf("Error(%v): status %d, want %d", tc.err, w.Code, tc.want)
		}
		if tc.want == http.StatusInternalServerError && strings.Contains(w.Body.String(), "10.0.0.5") {
			t.Errorf("internal error detail leaked to client: %s", w.Body)
		}
	}
}

type sampleRequest struct {
	FullName string `json:"full_name" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Age      int    `json:"age" binding:"gte=17"`
}

func newBindRouter(limit int64) *gin.Engine {
	r := gin.New()
	r.POST("/", func(c *gin.Context) {
		if limit > 0 {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		}
		var req sampleRequest
		if !BindJSON(c, &req) {
			return
		}
		c.Status(http.StatusNoContent)
	})
	return r
}

func post(r *gin.Engine, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestBindJSON_ValidationUsesJSONFieldNames(t *testing.T) {
	w := post(newBindRouter(0), `{"full_name":"Al","email":"not-an-email","age":10}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", w.Code)
	}

	var body response.Body
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"full_name": "must be at least 3 characters",
		"email":     "must be a valid email address",
		"age":       "must be greater than or equal to 17",
	}
	for field, msg := range want {
		if body.Errors[field] != msg {
			t.Errorf("errors[%q] = %q, want %q", field, body.Errors[field], msg)
		}
	}
}

func TestBindJSON_BadInput(t *testing.T) {
	if w := post(newBindRouter(0), `{bad json`); w.Code != http.StatusBadRequest {
		t.Errorf("malformed json: status %d, want 400", w.Code)
	}
	big := `{"full_name":"` + strings.Repeat("a", 200) + `","email":"a@b.co","age":20}`
	if w := post(newBindRouter(50), big); w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("too large: status %d, want 413", w.Code)
	}
	if w := post(newBindRouter(0), `{"full_name":"Alice","email":"a@b.co","age":20}`); w.Code != http.StatusNoContent {
		t.Errorf("valid body: status %d, want 204", w.Code)
	}
}
