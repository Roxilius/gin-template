// Package model berisi entity yang dipetakan GORM ke tabel database.
//
// Satu file per entity, contoh internal/model/product.go:
//
//	type Product struct {
//		ID        uint   `gorm:"primaryKey"`
//		Name      string `gorm:"size:100;not null"`
//		CreatedAt time.Time
//		UpdatedAt time.Time
//	}
//
// Model sengaja tidak diberi tag JSON. Yang dikirim ke client selalu DTO
// (internal/dto), supaya field sensitif tidak bocor dan bentuk API tidak
// terikat ke struktur tabel. Tabelnya dibuat lewat file migration (migrations/),
// bukan AutoMigrate.
package model
