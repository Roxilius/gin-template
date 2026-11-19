package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/Roxilius/gin-template/internal/domain"
)

// Translate memetakan error GORM ke error domain supaya service dan handler
// tidak bergantung pada detail database. gorm.ErrDuplicatedKey hanya muncul
// karena koneksi dibuka dengan TranslateError: true (lihat database/postgres.go).
func Translate(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return domain.ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return domain.ErrConflict
	default:
		return err
	}
}
