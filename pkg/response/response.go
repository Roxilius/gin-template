// Package response menyeragamkan format JSON response API.
package response

import "github.com/gin-gonic/gin"

// Body adalah amplop standar untuk semua response.
type Body struct {
	Success bool              `json:"success"`
	Message string            `json:"message,omitempty"`
	Data    any               `json:"data,omitempty"`
	Meta    any               `json:"meta,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
}

func Success(c *gin.Context, status int, data any) {
	c.JSON(status, Body{Success: true, Data: data})
}

func SuccessWithMeta(c *gin.Context, status int, data, meta any) {
	c.JSON(status, Body{Success: true, Data: data, Meta: meta})
}

func Error(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, Body{Success: false, Message: message})
}

func ValidationError(c *gin.Context, status int, fields map[string]string) {
	c.AbortWithStatusJSON(status, Body{Success: false, Message: "validation failed", Errors: fields})
}
