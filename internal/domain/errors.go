// Package domain berisi error bisnis generik yang dipakai bersama oleh semua layer.
//
// Alurnya:
//   - repository menerjemahkan error database ke error di sini
//     (mis. gorm.ErrRecordNotFound -> ErrNotFound),
//   - service mengembalikannya apa adanya atau membungkusnya dengan fmt.Errorf("...: %w", err),
//   - handler memetakannya ke status HTTP lewat handler.Error (lihat handler/errors.go).
//
// Tambahkan error spesifik project di sini bila perlu.
package domain

import "errors"

var (
	ErrNotFound     = errors.New("resource not found")
	ErrConflict     = errors.New("resource already exists")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)
