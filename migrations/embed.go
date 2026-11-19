// Package migrations menyimpan file SQL migration dan meng-embed-nya ke binary,
// sehingga binary migrate bisa dijalankan tanpa menyalin folder ini.
//
// Konvensi nama file (golang-migrate):
//
//	000001_create_products_table.up.sql
//	000001_create_products_table.down.sql
//
// Buat pasangan file baru dengan:
//
//	make migrate-create name=create_products_table
//	./scripts/dev.ps1 migrate-create create_products_table
package migrations

import "embed"

// FS berisi semua file di folder ini. Pola "*" (bukan "*.sql") dipakai supaya
// template tetap bisa di-compile walaupun belum ada file migration sama sekali;
// file yang bukan migration (seperti embed.go) otomatis diabaikan golang-migrate.
//
//go:embed *
var FS embed.FS
