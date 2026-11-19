package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SecureHeaders menambahkan header keamanan dasar untuk API JSON.
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		c.Next()
	}
}

// BodyLimit membatasi ukuran body request untuk mencegah client mengirim
// payload raksasa yang menghabiskan memori.
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}

// CORS mengizinkan origin tertentu (mis. frontend) memanggil API dari browser.
// Mengembalikan nil jika tidak ada origin yang dikonfigurasi (CORS nonaktif).
func CORS(allowedOrigins []string) gin.HandlerFunc {
	if len(allowedOrigins) == 0 {
		return nil
	}
	cfg := cors.Config{
		AllowMethods:  []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept", "Authorization", HeaderRequestID},
		ExposeHeaders: []string{HeaderRequestID},
		MaxAge:        12 * time.Hour,
	}
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		cfg.AllowAllOrigins = true
	} else {
		cfg.AllowOrigins = allowedOrigins
	}
	return cors.New(cfg)
}
