// Package repository bertanggung jawab atas akses data via GORM.
//
// Pola per entity, contoh internal/repository/product_repository.go:
//
//	type ProductRepository interface {
//		FindByID(ctx context.Context, id uint) (*model.Product, error)
//		Create(ctx context.Context, p *model.Product) error
//	}
//
//	type productRepository struct{ db *gorm.DB }
//
//	func NewProductRepository(db *gorm.DB) ProductRepository {
//		return &productRepository{db: db}
//	}
//
//	func (r *productRepository) FindByID(ctx context.Context, id uint) (*model.Product, error) {
//		var p model.Product
//		if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
//			return nil, Translate(err)
//		}
//		return &p, nil
//	}
//
// Selalu pakai db.WithContext(ctx) supaya query ikut dibatalkan saat request
// dibatalkan, dan selalu terjemahkan error lewat Translate (lihat errors.go).
package repository
