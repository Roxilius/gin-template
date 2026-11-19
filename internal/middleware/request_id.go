// Package middleware berisi middleware Gin yang dipakai semua route.
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"

	"github.com/gin-gonic/gin"
)

const (
	HeaderRequestID = "X-Request-ID"
	CtxRequestID    = "request_id"
)

// Hanya terima request ID dari client yang formatnya aman, supaya tidak bisa
// dipakai untuk menyisipkan karakter aneh ke log (log injection).
var validRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// RequestID memberi ID unik pada setiap request, berguna untuk menelusuri log
// satu request dari ujung ke ujung (juga lintas service).
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderRequestID)
		if !validRequestID.MatchString(id) {
			id = newID()
		}
		c.Set(CtxRequestID, id)
		c.Header(HeaderRequestID, id)
		c.Next()
	}
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
