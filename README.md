# Gin Template

Template REST API **Go + Gin + GORM + PostgreSQL** yang siap dipakai sebagai titik awal project baru.

Template ini sengaja **tidak berisi fitur bisnis atau entity apa pun**. Isinya hanya fondasi yang dibutuhkan hampir setiap project:

- Konfigurasi dari environment variable / `.env`, dengan validasi
- Koneksi PostgreSQL via GORM, termasuk connection pool dan ping saat startup
- Migration SQL versioned (golang-migrate), di-embed ke binary
- Structured logging (`log/slog`): teks untuk lokal, JSON untuk production
- Middleware: Request ID, logger, panic recovery, security headers, batas ukuran body, CORS
- Format response JSON yang seragam, dan pemetaan error ke status HTTP
- Pesan validasi yang memakai nama field JSON
- Health check `/health` (liveness) dan `/ready` (readiness + cek database)
- Graceful shutdown
- Swagger / OpenAPI
- Docker (multi-stage, distroless, non-root) dan docker-compose
- GitHub Actions CI (test, lint, build, docker build) dan golangci-lint
- Hot reload dengan `air`
- `Makefile` (Linux/macOS/CI) dan `scripts/dev.ps1` (Windows PowerShell)

---

## Daftar isi

- [Membuat project baru dari template ini](#membuat-project-baru-dari-template-ini)
- [Prasyarat](#prasyarat)
- [Quick start](#quick-start)
- [Struktur folder](#struktur-folder)
- [Endpoint bawaan](#endpoint-bawaan)
- [Perintah yang tersedia](#perintah-yang-tersedia)
- [Konfigurasi](#konfigurasi)
- [Database migration](#database-migration)
- [Menambah fitur baru (contoh: Product)](#menambah-fitur-baru-contoh-product)
- [Format response](#format-response)
- [Testing](#testing)
- [Docker](#docker)
- [Catatan keamanan](#catatan-keamanan)

---

## Membuat project baru dari template ini

Cukup tentukan **nama folder** project baru. Tidak perlu clone repo, tidak perlu akun GitHub, dan tidak perlu tool tambahan selain Go.

### Langkah

```bash
# 1. Pindah ke folder tempat project akan dibuat
cd ~/projects                       # PowerShell: cd C:\path\ke\folder-project

# 2. Buat project baru
go run github.com/Roxilius/gin-template/scripts/initproject@main new-project

# 3. Masuk ke project baru
cd new-project

# 4. Jalankan PostgreSQL (butuh Docker Desktop)
make db-up                          # PowerShell: ./scripts/dev.ps1 db-up

# 5. Jalankan API
make run                            # PowerShell: ./scripts/dev.ps1 run
```

Buka http://localhost:8080/swagger untuk melihat dokumentasi API.

Ganti **`new-project`** dengan nama project kamu. Contoh, untuk project bernama `shop-api`:

```bash
go run github.com/Roxilius/gin-template/scripts/initproject@main shop-api
cd shop-api
```

Bagian `github.com/Roxilius/gin-template/scripts/initproject@main` adalah alamat template ini dan **selalu sama**, jadi jangan diganti.

> **Tanpa Docker?** Pakai PostgreSQL yang sudah ter-install di laptop. Lewati langkah 4, lalu sesuaikan `DB_USER`, `DB_PASSWORD`, dan `DB_NAME` di file `.env` sebelum langkah 5. Aplikasi langsung berhenti jika tidak bisa terhubung ke database.

> **Pengguna Windows:** jika muncul error *"running scripts is disabled on this system"*, jalankan sekali `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned`.

### Apa yang dilakukan perintah langkah 2

Contoh di bawah memakai nama project `shop-api`.

1. Mengunduh versi terbaru template, lalu menyalinnya ke folder baru `shop-api/`. Isinya bersih: tanpa `.git`, `.env`, atau file build.
2. Memakai nama folder sebagai nama project:
   - module di `go.mod` dan semua import → `shop-api`
   - `APP_NAME` → `shop-api`
   - judul Swagger dan README → `Shop API`
3. Membuat file `.env` dari `.env.example`.

Nama folder hanya boleh berisi huruf, angka, `-`, `_`, dan `.` (tanpa spasi). Perintah ini menolak jika folder tujuan sudah ada dan berisi file.

Opsi tambahan (taruh **sebelum** nama folder):

| Opsi | Contoh | Fungsi |
|---|---|---|
| `-title` | `-title "Toko Kita API"` | Judul Swagger selain yang dibuat otomatis dari nama folder |
| `-name` | `-name toko` | `APP_NAME` yang berbeda dari nama folder |
| `-module` | `-module github.com/nama/shop-api` | Module path lengkap, jika project ini nanti mau di-import oleh project Go lain |

### Hot reload (opsional)

Agar API otomatis restart setiap file disimpan, install `air` sekali, lalu pakai `dev`, bukan `run`:

```bash
go install github.com/air-verse/air@latest
make dev                            # PowerShell: ./scripts/dev.ps1 dev
```

### Setelah project dibuat

- Sesuaikan `@description` Swagger di [cmd/api/main.go](cmd/api/main.go) dan isi README sesuai project.
- Folder `scripts/initproject` boleh dihapus, karena hanya dipakai saat membuat project.
- Project yang sudah dibuat **tidak ikut berubah** jika template ini diperbarui. Hanya project baru yang mendapat versi terbaru.

### Cara lain: salin manual

Jika sudah punya salinan folder template, misalnya dari `git clone` atau file ZIP:

1. Ganti nama foldernya menjadi nama project, misalnya `shop-api`, lalu hapus folder `.git` di dalamnya jika ada.
2. Dari dalam folder itu, jalankan `make init` / `./scripts/dev.ps1 init`. Hasilnya sama dengan perintah di atas.

---

## Prasyarat

| Tool | Keterangan |
|---|---|
| Go 1.27+ | Wajib |
| Docker Desktop | Untuk menjalankan PostgreSQL atau seluruh stack |
| `make` | Opsional. Di Windows pakai `./scripts/dev.ps1` |
| air, swag, golangci-lint | Opsional. Install dengan `make tools` / `./scripts/dev.ps1 tools` |

---

## Quick start

```bash
# 1. Siapkan environment
cp .env.example .env                # PowerShell: Copy-Item .env.example .env

# 2. Jalankan PostgreSQL
make db-up                          # PowerShell: ./scripts/dev.ps1 db-up

# 3. Jalankan migration (kalau sudah ada file migration)
make migrate-up                     # PowerShell: ./scripts/dev.ps1 migrate-up

# 4. Jalankan API
make run                            # PowerShell: ./scripts/dev.ps1 run
```

Lalu buka:

- http://localhost:8080/health
- http://localhost:8080/ready
- http://localhost:8080/api/v1/ping
- http://localhost:8080/swagger/index.html (jika `SWAGGER_ENABLED=true`)

---

## Struktur folder

```
.
├── cmd/
│   ├── api/main.go            # Entry point HTTP server: wiring dependency + graceful shutdown
│   └── migrate/main.go        # CLI migration: up / down / version / force / create
├── internal/                  # Kode aplikasi (tidak bisa di-import project lain)
│   ├── config/                # Load & validasi konfigurasi dari env
│   ├── database/              # Koneksi PostgreSQL (GORM)
│   ├── domain/                # Error bisnis generik (ErrNotFound, ErrConflict, ...)
│   ├── model/                 # Entity GORM            (kosong, isi per project)
│   ├── dto/                   # Request/response API   (kosong, isi per project)
│   ├── repository/            # Akses data + Translate error GORM -> domain
│   ├── service/               # Business logic         (kosong, isi per project)
│   ├── handler/               # Layer HTTP + helper Error/BindJSON/BindQuery
│   ├── middleware/            # Request ID, logger, recovery, security, CORS
│   └── router/                # Registrasi middleware & route
├── pkg/                       # Kode generik yang boleh dipakai ulang
│   ├── logger/                # Setup slog
│   └── response/              # Format JSON response standar
├── migrations/                # File SQL migration (di-embed ke binary)
├── docs/                      # Swagger (di-generate, jangan diedit manual)
├── scripts/dev.ps1            # Task runner untuk Windows
├── .github/workflows/ci.yml   # CI pipeline
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── .air.toml                  # Konfigurasi hot reload
├── .golangci.yml              # Konfigurasi linter
└── .env.example
```

Setiap folder yang masih kosong (`model`, `dto`, `repository`, `service`) punya file `doc.go` berisi contoh kode untuk folder tersebut.

### Arah dependency

```
handler  ──►  service  ──►  repository  ──►  model
   │             │               │
   └─────────────┴───────────────┴──►  domain (error bersama)
```

- **handler** hanya mengurus HTTP: parsing, validasi, dan menulis response. Tidak ada business logic di sini.
- **service** berisi business logic dan tidak boleh meng-import Gin.
- **repository** satu-satunya layer yang menyentuh database.
- Layer bawah **tidak boleh** meng-import layer di atasnya.

---

## Endpoint bawaan

| Method | Path | Keterangan |
|---|---|---|
| GET | `/health` | Liveness: proses hidup. Tidak mengecek dependency. |
| GET | `/ready` | Readiness: termasuk ping database. Mengembalikan `503` jika database mati. |
| GET | `/api/v1/ping` | Endpoint test, membalas `pong`. |
| GET | `/swagger/*` | Dokumentasi API (hanya jika `SWAGGER_ENABLED=true`). |

Contoh `/ready`:

```json
{
  "success": true,
  "data": {
    "status": "ok",
    "checks": { "database": "up" },
    "time": "2026-11-19T15:00:00Z",
    "duration": "1.2ms"
  }
}
```

---

## Perintah yang tersedia

| Make | PowerShell | Fungsi |
|---|---|---|
| `make init` | `./scripts/dev.ps1 init` | Sesuaikan nama project hasil salinan manual template |
| `make run` | `./scripts/dev.ps1 run` | Jalankan API |
| `make dev` | `./scripts/dev.ps1 dev` | Jalankan dengan hot reload (air) |
| `make build` | `./scripts/dev.ps1 build` | Build binary ke `./bin` |
| `make test` | `./scripts/dev.ps1 test` | Jalankan semua test |
| `make cover` | `./scripts/dev.ps1 cover` | Test + `coverage.html` |
| `make lint` | `./scripts/dev.ps1 lint` | golangci-lint |
| `make swag` | `./scripts/dev.ps1 swag` | Generate ulang Swagger |
| `make migrate-up` | `./scripts/dev.ps1 migrate-up` | Jalankan migration |
| `make migrate-down n=1` | `./scripts/dev.ps1 migrate-down 1` | Rollback migration |
| `make migrate-create name=x` | `./scripts/dev.ps1 migrate-create x` | Buat file migration baru |
| `make db-up` | `./scripts/dev.ps1 db-up` | Jalankan PostgreSQL saja |
| `make docker-up` | `./scripts/dev.ps1 docker-up` | Jalankan seluruh stack |
| `make docker-down` | `./scripts/dev.ps1 docker-down` | Hentikan stack |
| `make tools` | `./scripts/dev.ps1 tools` | Install air, swag, golangci-lint |

> Jika PowerShell menolak menjalankan script, jalankan sekali:
> `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned`

---

## Konfigurasi

Semua konfigurasi dibaca dari environment variable. File `.env` dimuat otomatis jika ada, tapi **tidak menimpa** env yang sudah di-set. Daftar lengkapnya ada di [.env.example](.env.example).

| Variable | Default | Keterangan |
|---|---|---|
| `APP_ENV` | `development` | `development` / `test` / `production` |
| `APP_PORT` | `8080` | Port HTTP |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `SWAGGER_ENABLED` | `false` | Aktifkan `/swagger` |
| `CORS_ALLOWED_ORIGINS` | *(kosong)* | Daftar origin dipisah koma. Kosong berarti CORS nonaktif. |
| `TRUSTED_PROXIES` | *(kosong)* | IP/CIDR reverse proxy yang dipercaya |
| `HTTP_MAX_BODY_BYTES` | `1048576` | Ukuran maksimum body request (1 MB) |
| `DB_HOST` / `DB_PORT` | `localhost` / `5432` | Lokasi PostgreSQL |
| `DB_USER` / `DB_PASSWORD` / `DB_NAME` | `app` / *(kosong)* / `app` | Kredensial database |
| `DB_SSLMODE` | `disable` | Wajib selain `disable` di production |

Saat `APP_ENV=production`, aplikasi **menolak start** jika `DB_PASSWORD` kosong atau `DB_SSLMODE=disable`.

Untuk menambah konfigurasi baru (misalnya Redis): tambahkan field di struct [internal/config/config.go](internal/config/config.go), isi di `Load()`, validasi di `validate()`, lalu dokumentasikan di `.env.example`.

---

## Database migration

Migration memakai file SQL biasa di folder [migrations/](migrations/), dengan format nama dari [golang-migrate](https://github.com/golang-migrate/migrate):

```
000001_create_products_table.up.sql     # perubahan
000001_create_products_table.down.sql   # cara membatalkannya
```

```bash
make migrate-create name=create_products_table   # buat pasangan file baru
make migrate-up                                  # jalankan semua yang belum
make migrate-down                                # rollback 1 langkah
make migrate-version                             # cek versi saat ini
go run ./cmd/migrate force 1                     # perbaiki state "dirty" setelah migration gagal
```

File migration di-embed ke binary `migrate`, jadi di server atau Docker tidak perlu menyalin folder `migrations/`.

> Template ini sengaja **tidak memakai `AutoMigrate`** GORM. Migration berversi lebih aman untuk production: bisa direview, di-rollback, dan hasilnya sama di semua environment.

---

## Menambah fitur baru (contoh: Product)

Urutannya selalu dari bawah ke atas: **migration → model → repository → dto → service → handler → router → main**.

**1. Migration**

```bash
make migrate-create name=create_products_table
```

```sql
-- migrations/000001_create_products_table.up.sql
CREATE TABLE products (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    price      BIGINT       NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- migrations/000001_create_products_table.down.sql
DROP TABLE IF EXISTS products;
```

**2. Model** (`internal/model/product.go`)

```go
type Product struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:100;not null"`
	Price     int64  `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
```

**3. Repository** (`internal/repository/product_repository.go`)

```go
type ProductRepository interface {
	FindByID(ctx context.Context, id uint) (*model.Product, error)
	Create(ctx context.Context, p *model.Product) error
}

type productRepository struct{ db *gorm.DB }

func NewProductRepository(db *gorm.DB) ProductRepository { return &productRepository{db: db} }

func (r *productRepository) FindByID(ctx context.Context, id uint) (*model.Product, error) {
	var p model.Product
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, Translate(err) // gorm.ErrRecordNotFound -> domain.ErrNotFound
	}
	return &p, nil
}

func (r *productRepository) Create(ctx context.Context, p *model.Product) error {
	return Translate(r.db.WithContext(ctx).Create(p).Error)
}
```

**4. DTO** (`internal/dto/product.go`)

```go
type CreateProductRequest struct {
	Name  string `json:"name" binding:"required,min=2,max=100"`
	Price int64  `json:"price" binding:"required,gte=0"`
}

type ProductResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Price int64  `json:"price"`
}

func NewProductResponse(p *model.Product) ProductResponse {
	return ProductResponse{ID: p.ID, Name: p.Name, Price: p.Price}
}
```

**5. Service** (`internal/service/product_service.go`)

```go
type ProductService interface {
	Get(ctx context.Context, id uint) (*model.Product, error)
	Create(ctx context.Context, req dto.CreateProductRequest) (*model.Product, error)
}

type productService struct{ repo repository.ProductRepository }

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) Get(ctx context.Context, id uint) (*model.Product, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *productService) Create(ctx context.Context, req dto.CreateProductRequest) (*model.Product, error) {
	p := &model.Product{Name: strings.TrimSpace(req.Name), Price: req.Price}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}
```

**6. Handler** (`internal/handler/product_handler.go`)

```go
type ProductHandler struct{ svc service.ProductService }

func NewProductHandler(svc service.ProductService) *ProductHandler { return &ProductHandler{svc: svc} }

// Create godoc
//
//	@Summary	Create product
//	@Tags		products
//	@Accept		json
//	@Produce	json
//	@Param		request	body		dto.CreateProductRequest	true	"Payload"
//	@Success	201		{object}	response.Body{data=dto.ProductResponse}
//	@Failure	400		{object}	response.Body
//	@Router		/api/v1/products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var req dto.CreateProductRequest
	if !BindJSON(c, &req) { // response 400 + detail per field sudah ditulis otomatis
		return
	}
	p, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		Error(c, err) // domain error -> 404/409/..., error lain -> 500 generik
		return
	}
	response.Success(c, http.StatusCreated, dto.NewProductResponse(p))
}
```

**7. Router** ([internal/router/router.go](internal/router/router.go)): tambahkan handler ke `Deps`, lalu daftarkan route-nya.

```go
type Deps struct {
	// ...
	Product *handler.ProductHandler
}

products := v1.Group("/products")
products.GET("/:id", d.Product.Get)
products.POST("", d.Product.Create)
```

**8. Wiring** ([cmd/api/main.go](cmd/api/main.go))

```go
productRepo := repository.NewProductRepository(db)
productSvc := service.NewProductService(productRepo)
productH := handler.NewProductHandler(productSvc)

r, err := router.New(router.Deps{Config: cfg, Logger: log, Health: healthH, Product: productH})
```

**9.** Jalankan `make migrate-up`, `make swag`, dan `make test`.

---

## Format response

Semua response memakai amplop yang sama ([pkg/response](pkg/response/response.go)):

```jsonc
// Sukses
{ "success": true, "data": { ... }, "meta": { ... } }

// Error
{ "success": false, "message": "resource not found" }

// Error validasi (400)
{
  "success": false,
  "message": "validation failed",
  "errors": { "name": "is required", "price": "must be greater than or equal to 0" }
}
```

Pemetaan error di `handler.Error`:

| Error | HTTP |
|---|---|
| `domain.ErrNotFound` | 404 |
| `domain.ErrConflict` | 409 |
| `domain.ErrUnauthorized` | 401 |
| `domain.ErrForbidden` | 403 |
| Error lain | 500 dengan pesan generik. Detailnya hanya dicatat di log. |

Error dari layer bawah boleh dibungkus (`fmt.Errorf("create product: %w", err)`). Pemetaannya tetap berfungsi karena memakai `errors.Is`.

---

## Testing

```bash
make test     # semua test
make cover    # dengan laporan coverage
```

- **Unit test service**: buat fake repository in-memory yang memenuhi interface repository, lalu suntikkan ke `NewXxxService`. Dengan begitu service bisa dites tanpa database.
- **Test HTTP**: lihat contohnya di [internal/router/router_test.go](internal/router/router_test.go), yang memakai `httptest` tanpa menjalankan server sungguhan.
- CI menjalankan test dengan `-race` untuk mendeteksi data race.

---

## Docker

```bash
cp .env.example .env
docker compose up -d --build     # db -> migrate (sekali jalan) -> api
docker compose logs -f api
docker compose down              # tambah -v untuk menghapus data database
```

Image akhir memakai `distroless/static:nonroot`: tanpa shell, berjalan sebagai user non-root, dan ukurannya kecil. Binary `migrate` ikut ada di dalam image:

```bash
docker compose run --rm migrate          # jalankan migration
docker compose run --rm --entrypoint /app/migrate api version
```

---

## Catatan keamanan

Yang sudah ditangani oleh template:

- Error internal tidak pernah dikirim ke client, hanya dicatat di log.
- Query string tidak ikut dicatat di log, karena bisa berisi token.
- `X-Request-ID` dari client hanya diterima jika formatnya aman, untuk mencegah log injection.
- `TRUSTED_PROXIES` kosong secara default, sehingga `ClientIP()` tidak bisa dipalsukan lewat header `X-Forwarded-For`.
- Ukuran body request dibatasi (default 1 MB).
- Tersedia header keamanan dasar dan timeout HTTP server, termasuk perlindungan dari serangan Slowloris.
- Production menolak start tanpa password database atau tanpa SSL.
- Container berjalan sebagai user non-root.

Yang perlu ditambahkan sendiri sesuai kebutuhan project:

- Autentikasi/otorisasi (JWT, session, OAuth)
- Rate limiting, terutama untuk endpoint login
- HTTPS/TLS, biasanya di reverse proxy atau load balancer
- Secret management di production (bukan file `.env`)
