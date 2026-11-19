package handler

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/Roxilius/gin-template/internal/domain"
	"github.com/Roxilius/gin-template/pkg/response"
)

// Error menulis response error berdasarkan jenis error domain. Error yang tidak
// dikenal dicatat (via c.Error -> middleware Logger) dan dibalas pesan generik,
// supaya detail internal seperti query SQL tidak bocor ke client.
func Error(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		response.Error(c, http.StatusNotFound, domain.ErrNotFound.Error())
	case errors.Is(err, domain.ErrConflict):
		response.Error(c, http.StatusConflict, domain.ErrConflict.Error())
	case errors.Is(err, domain.ErrUnauthorized):
		response.Error(c, http.StatusUnauthorized, domain.ErrUnauthorized.Error())
	case errors.Is(err, domain.ErrForbidden):
		response.Error(c, http.StatusForbidden, domain.ErrForbidden.Error())
	default:
		_ = c.Error(err)
		response.Error(c, http.StatusInternalServerError, "internal server error")
	}
}

// BindJSON mem-parsing body JSON ke obj dan menjalankan validasi tag `binding`.
// Jika gagal, response 400/413 sudah ditulis dan fungsi mengembalikan false:
//
//	var req dto.CreateProductRequest
//	if !handler.BindJSON(c, &req) {
//		return
//	}
func BindJSON(c *gin.Context, obj any) bool {
	return bind(c, c.ShouldBindJSON(obj))
}

// BindQuery sama seperti BindJSON, tapi untuk query string (?page=1).
func BindQuery(c *gin.Context, obj any) bool {
	return bind(c, c.ShouldBindQuery(obj))
}

func bind(c *gin.Context, err error) bool {
	if err == nil {
		return true
	}

	var verrs validator.ValidationErrors
	var tooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &verrs):
		fields := make(map[string]string, len(verrs))
		for _, fe := range verrs {
			fields[fe.Field()] = validationMessage(fe)
		}
		response.ValidationError(c, http.StatusBadRequest, fields)
	case errors.As(err, &tooLarge):
		response.Error(c, http.StatusRequestEntityTooLarge, "request body too large")
	default:
		response.Error(c, http.StatusBadRequest, "invalid request body")
	}
	return false
}

func validationMessage(fe validator.FieldError) string {
	unit := ""
	if fe.Kind() == reflect.String {
		unit = " characters"
	}
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s%s", fe.Param(), unit)
	case "max":
		return fmt.Sprintf("must be at most %s%s", fe.Param(), unit)
	case "len":
		return fmt.Sprintf("must be exactly %s%s", fe.Param(), unit)
	case "gte":
		return "must be greater than or equal to " + fe.Param()
	case "lte":
		return "must be less than or equal to " + fe.Param()
	case "oneof":
		return "must be one of: " + fe.Param()
	default:
		return "is invalid"
	}
}

var registerOnce sync.Once

// RegisterValidator membuat pesan validasi memakai nama field JSON
// ("first_name") alih-alih nama field Go ("FirstName"). Dipanggil oleh router.
func RegisterValidator() {
	registerOnce.Do(func() {
		v, ok := binding.Validator.Engine().(*validator.Validate)
		if !ok {
			return
		}
		v.RegisterTagNameFunc(func(f reflect.StructField) string {
			for _, tag := range []string{"json", "form"} {
				name, _, _ := strings.Cut(f.Tag.Get(tag), ",")
				if name == "-" {
					return ""
				}
				if name != "" {
					return name
				}
			}
			return f.Name
		})
	})
}
