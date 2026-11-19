package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Roxilius/gin-template/pkg/response"
)

// Logger mencatat setiap request dalam format structured log (slog).
// Query string sengaja tidak dicatat karena bisa berisi data sensitif.
func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		attrs := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency", time.Since(start).String(),
			"ip", c.ClientIP(),
			"request_id", c.GetString(CtxRequestID),
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, "errors", c.Errors.String())
		}

		level := slog.LevelInfo
		switch {
		case status >= http.StatusInternalServerError:
			level = slog.LevelError
		case status >= http.StatusBadRequest:
			level = slog.LevelWarn
		}
		log.Log(c.Request.Context(), level, "http request", attrs...)
	}
}

// Recovery menangkap panic, mencatatnya, dan membalas JSON 500 standar
// supaya server tidak mati dan stack trace tidak terkirim ke client.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered any) {
		log.Error("panic recovered",
			"error", recovered,
			"path", c.Request.URL.Path,
			"request_id", c.GetString(CtxRequestID),
		)
		response.Error(c, http.StatusInternalServerError, "internal server error")
	})
}
