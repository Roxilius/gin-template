// Command api menjalankan HTTP server.
//
// Tugas main hanya "merakit" dependency (wiring) lalu menjalankan server
// dengan graceful shutdown. Logika apa pun ada di package internal/.
//
//	@title						Gin Template API
//	@version					1.0
//	@description				Template REST API berbasis Gin + GORM + PostgreSQL.
//	@BasePath					/
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Format: "Bearer <token>"
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	_ "github.com/Roxilius/gin-template/docs" // registrasi spesifikasi Swagger
	"github.com/Roxilius/gin-template/internal/config"
	"github.com/Roxilius/gin-template/internal/database"
	"github.com/Roxilius/gin-template/internal/handler"
	"github.com/Roxilius/gin-template/internal/router"
	"github.com/Roxilius/gin-template/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application stopped with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.App.Env, cfg.Log.Level)
	slog.SetDefault(log)
	gin.SetMode(ginMode(cfg.App.Env))

	// --- Infrastructure ---
	db, err := database.NewPostgres(cfg.DB, cfg.App.Env == config.EnvDevelopment)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer func() { _ = sqlDB.Close() }()
	log.Info("database connected", "host", cfg.DB.Host, "name", cfg.DB.Name)

	// --- Wiring: repository -> service -> handler ---
	// Contoh saat menambah fitur:
	//   productRepo := repository.NewProductRepository(db)
	//   productSvc  := service.NewProductService(productRepo)
	//   productH    := handler.NewProductHandler(productSvc)
	healthH := handler.NewHealthHandler(cfg.App.Name, sqlDB.PingContext)

	r, err := router.New(router.Deps{Config: cfg, Logger: log, Health: healthH})
	if err != nil {
		return fmt.Errorf("build router: %w", err)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.HTTP.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("http server started", "addr", srv.Addr, "env", cfg.App.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Tunggu Ctrl+C / SIGTERM (dikirim Docker & Kubernetes saat stop).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
	}

	log.Info("shutting down server", "timeout", cfg.App.ShutdownTimeout.String())
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	log.Info("server stopped")
	return nil
}

func ginMode(env string) string {
	switch env {
	case config.EnvProduction:
		return gin.ReleaseMode
	case config.EnvTest:
		return gin.TestMode
	default:
		return gin.DebugMode
	}
}
