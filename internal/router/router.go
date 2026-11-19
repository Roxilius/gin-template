// Package router mendaftarkan semua middleware dan route.
package router

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/Roxilius/gin-template/internal/config"
	"github.com/Roxilius/gin-template/internal/handler"
	"github.com/Roxilius/gin-template/internal/middleware"
	"github.com/Roxilius/gin-template/pkg/response"
)

// Deps berisi semua dependency yang dibutuhkan router.
// Saat menambah fitur, tambahkan handler-nya di sini.
type Deps struct {
	Config config.Config
	Logger *slog.Logger
	Health *handler.HealthHandler
}

func New(d Deps) (*gin.Engine, error) {
	handler.RegisterValidator()

	r := gin.New()

	// nil = tidak percaya proxy mana pun, sehingga c.ClientIP() tidak bisa
	// dipalsukan lewat header X-Forwarded-For. Isi TRUSTED_PROXIES jika
	// aplikasi berada di belakang load balancer / reverse proxy.
	if err := r.SetTrustedProxies(d.Config.HTTP.TrustedProxies); err != nil {
		return nil, err
	}

	r.Use(
		middleware.RequestID(),
		middleware.Logger(d.Logger),
		middleware.Recovery(d.Logger),
		middleware.SecureHeaders(),
		middleware.BodyLimit(d.Config.HTTP.MaxBodyBytes),
	)
	if corsMw := middleware.CORS(d.Config.HTTP.CORSAllowedOrigins); corsMw != nil {
		r.Use(corsMw)
	}

	// Health check untuk Docker / Kubernetes / load balancer.
	r.GET("/health", d.Health.Live)
	r.GET("/ready", d.Health.Ready)

	if d.Config.Swagger.Enabled {
		swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
		r.GET("/swagger/*any", func(c *gin.Context) {
			// gin-swagger hanya melayani nama file tertentu (index.html, doc.json,
			// ...), sehingga /swagger/ akan 404. Arahkan ke halaman UI-nya.
			if c.Param("any") == "/" {
				c.Redirect(http.StatusFound, "/swagger/index.html")
				return
			}
			swaggerHandler(c)
		})
	}

	// Semua endpoint bisnis masuk ke group versi, mis. /api/v1/products.
	v1 := r.Group("/api/v1")
	v1.GET("/ping", d.Health.Ping)
	// Contoh menambah fitur:
	//   products := v1.Group("/products")
	//   products.GET("", d.Product.List)
	//   products.POST("", d.Product.Create)

	r.NoRoute(func(c *gin.Context) { response.Error(c, http.StatusNotFound, "route not found") })
	r.HandleMethodNotAllowed = true
	r.NoMethod(func(c *gin.Context) { response.Error(c, http.StatusMethodNotAllowed, "method not allowed") })

	return r, nil
}
