// Package database membuka koneksi PostgreSQL melalui GORM.
package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/Roxilius/gin-template/internal/config"
)

// NewPostgres membuka koneksi, mengatur connection pool, lalu memastikan
// database benar-benar bisa dijangkau (ping) sebelum aplikasi lanjut.
func NewPostgres(cfg config.DBConfig, logSQL bool) (*gorm.DB, error) {
	level := gormlogger.Warn
	if logSQL {
		level = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		// Ubah error spesifik database (mis. unique violation) menjadi
		// error GORM generik seperti gorm.ErrDuplicatedKey.
		TranslateError: true,
		Logger:         gormlogger.Default.LogMode(level),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}
