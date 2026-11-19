// Package dto berisi struct request/response untuk layer HTTP.
//
// Contoh internal/dto/product.go:
//
//	type CreateProductRequest struct {
//		Name string `json:"name" binding:"required,min=2,max=100"`
//	}
//
//	type ProductResponse struct {
//		ID   uint   `json:"id"`
//		Name string `json:"name"`
//	}
//
//	func NewProductResponse(p *model.Product) ProductResponse { ... }
//
// Tag `binding` divalidasi otomatis oleh handler.BindJSON / handler.BindQuery,
// dan pesan errornya memakai nama field JSON.
package dto
