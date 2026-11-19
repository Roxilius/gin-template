// Package service berisi business logic. Tidak boleh bergantung pada Gin/HTTP,
// supaya mudah dites dan bisa dipakai ulang (mis. dari worker atau CLI).
//
// Contoh internal/service/product_service.go:
//
//	type ProductService interface {
//		Create(ctx context.Context, req dto.CreateProductRequest) (*model.Product, error)
//	}
//
//	type productService struct{ repo repository.ProductRepository }
//
//	func NewProductService(repo repository.ProductRepository) ProductService {
//		return &productService{repo: repo}
//	}
//
// Untuk unit test, buat fake repository in-memory yang memenuhi interface
// repository, lalu suntikkan ke NewProductService.
package service
