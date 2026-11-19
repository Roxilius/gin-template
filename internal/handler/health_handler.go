// Package handler adalah layer HTTP: parsing request, validasi input,
// memanggil service, dan menulis response. Tidak berisi business logic.
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Roxilius/gin-template/pkg/response"
)

// PingFunc memeriksa dependency (mis. database). Dibuat sebagai fungsi supaya
// handler tidak bergantung langsung ke *gorm.DB dan mudah dites.
type PingFunc func(ctx context.Context) error

type HealthHandler struct {
	appName string
	pingDB  PingFunc
}

func NewHealthHandler(appName string, pingDB PingFunc) *HealthHandler {
	return &HealthHandler{appName: appName, pingDB: pingDB}
}

type HealthResponse struct {
	Status   string            `json:"status" example:"ok"`
	App      string            `json:"app,omitempty" example:"gin-template"`
	Checks   map[string]string `json:"checks,omitempty"`
	Time     time.Time         `json:"time"`
	Duration string            `json:"duration,omitempty" example:"1.2ms"`
}

// Live godoc
//
//	@Summary		Liveness probe
//	@Description	Mengecek proses aplikasi hidup. Tidak mengecek dependency.
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	response.Body{data=HealthResponse}
//	@Router			/health [get]
func (h *HealthHandler) Live(c *gin.Context) {
	response.Success(c, http.StatusOK, HealthResponse{Status: "ok", App: h.appName, Time: time.Now().UTC()})
}

// Ready godoc
//
//	@Summary		Readiness probe
//	@Description	Mengecek aplikasi siap menerima traffic, termasuk koneksi database.
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	response.Body{data=HealthResponse}
//	@Failure		503	{object}	response.Body{data=HealthResponse}
//	@Router			/ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	status, code, dbStatus := "ok", http.StatusOK, "up"
	if err := h.pingDB(ctx); err != nil {
		_ = c.Error(err) // dicatat oleh middleware Logger, tidak dikirim ke client
		status, code, dbStatus = "unavailable", http.StatusServiceUnavailable, "down"
	}

	body := HealthResponse{
		Status:   status,
		Checks:   map[string]string{"database": dbStatus},
		Time:     time.Now().UTC(),
		Duration: time.Since(start).String(),
	}
	c.JSON(code, response.Body{Success: code == http.StatusOK, Data: body})
}

// Ping godoc
//
//	@Summary	Test endpoint
//	@Tags		health
//	@Produce	json
//	@Success	200	{object}	response.Body{data=map[string]string}
//	@Router		/api/v1/ping [get]
func (h *HealthHandler) Ping(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{"message": "pong"})
}
