// Command migrate menjalankan database migration (golang-migrate) memakai file
// SQL yang di-embed dari folder migrations/.
//
// Pemakaian:
//
//	go run ./cmd/migrate up            # jalankan semua migration yang belum
//	go run ./cmd/migrate down [N]      # rollback N migration (default 1)
//	go run ./cmd/migrate version       # tampilkan versi saat ini
//	go run ./cmd/migrate force V       # paksa set versi (memperbaiki state "dirty")
//	go run ./cmd/migrate create NAME   # buat pasangan file .up.sql/.down.sql baru
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // driver "pgx5://"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/Roxilius/gin-template/internal/config"
	"github.com/Roxilius/gin-template/migrations"
)

const usage = `usage: migrate <command> [arg]

commands:
  up            apply all pending migrations
  down [N]      roll back N migrations (default 1)
  version       print current migration version
  force V       set version V without running migrations (fix dirty state)
  create NAME   create a new pair of migration files in ./migrations`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New(usage)
	}
	cmd, arg := args[0], ""
	if len(args) > 1 {
		arg = args[1]
	}

	// "create" hanya membuat file, tidak butuh koneksi database.
	if cmd == "create" {
		return create("migrations", arg)
	}

	if !hasMigrations() {
		fmt.Println("no migration files found in migrations/, nothing to do")
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("open migration source: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, cfg.DB.MigrateURL())
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	switch cmd {
	case "up":
		err = m.Up()
	case "down":
		n := 1
		if arg != "" {
			if n, err = strconv.Atoi(arg); err != nil || n < 1 {
				return errors.New("down: N must be a positive integer")
			}
		}
		err = m.Steps(-n)
	case "version":
		v, dirty, verr := m.Version()
		if errors.Is(verr, migrate.ErrNilVersion) {
			fmt.Println("no migration applied yet")
			return nil
		}
		if verr != nil {
			return verr
		}
		fmt.Printf("version: %d (dirty: %t)\n", v, dirty)
		return nil
	case "force":
		v, perr := strconv.Atoi(arg)
		if perr != nil {
			return errors.New("force: V must be an integer")
		}
		err = m.Force(v)
	default:
		return fmt.Errorf("unknown command %q\n\n%s", cmd, usage)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("no change")
		return nil
	}
	if err != nil {
		return err
	}
	fmt.Printf("%s: done\n", cmd)
	return nil
}

func hasMigrations() bool {
	files, _ := fs.Glob(migrations.FS, "*.up.sql")
	return len(files) > 0
}

var validName = regexp.MustCompile(`^[a-z0-9_]+$`)
var seqPrefix = regexp.MustCompile(`^(\d+)_`)

// create membuat file migration berikutnya dengan nomor urut 6 digit.
func create(dir, name string) error {
	if !validName.MatchString(name) {
		return errors.New("create: NAME is required and may only contain a-z, 0-9 and _ (e.g. create_products_table)")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read %s (run this from the project root): %w", dir, err)
	}
	next := 1
	for _, e := range entries {
		if m := seqPrefix.FindStringSubmatch(e.Name()); m != nil {
			if n, _ := strconv.Atoi(m[1]); n >= next {
				next = n + 1
			}
		}
	}

	base := fmt.Sprintf("%06d_%s", next, name)
	for _, suffix := range []string{".up.sql", ".down.sql"} {
		path := filepath.Join(dir, base+suffix)
		//nolint:gosec // file SQL adalah source code biasa, memang boleh dibaca (0644)
		if err := os.WriteFile(path, []byte("-- "+base+suffix+"\n"), 0o644); err != nil {
			return err
		}
		fmt.Println("created", path)
	}
	return nil
}
