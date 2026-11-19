// Command initproject menyesuaikan nama aplikasi setelah template disalin
// dengan gonew. gonew sudah mengganti module path di go.mod dan import Go;
// perintah ini mengganti sisanya (APP_NAME, judul Swagger, dll.).
//
// Jalankan sekali dari root project baru:
//
//	go run ./scripts/initproject                          # nama diambil dari go.mod
//	go run ./scripts/initproject -name shop-api -title "Shop API"
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	templateName  = "gin-template"
	templateTitle = "Gin Template API"
)

func main() {
	name := flag.String("name", "", "nama aplikasi (default: segmen terakhir module path di go.mod)")
	title := flag.String("title", "", `judul Swagger (default: diturunkan dari nama, mis. "Shop API")`)
	flag.Parse()

	changed, err := run(".", *name, *title)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if len(changed) == 0 {
		fmt.Println("Tidak ada yang perlu diubah (mungkin sudah pernah dijalankan).")
		return
	}
	fmt.Println("File yang diubah:")
	for _, f := range changed {
		fmt.Println("  -", f)
	}
}

// run mengganti nama template di root dan mengembalikan daftar file yang berubah.
func run(root, name, title string) ([]string, error) {
	if name == "" {
		module, err := readModulePath(filepath.Join(root, "go.mod"))
		if err != nil {
			return nil, err
		}
		name = nameFromModule(module)
	}
	if name == templateName {
		return nil, errors.New("nama project masih " + templateName +
			"; jalankan perintah ini di project baru hasil gonew, atau beri -name")
	}
	if title == "" {
		title = titleFromName(name)
	}

	appNameKV := "APP_NAME=" + templateName
	edits := []struct {
		path         string
		replacements []string // pasangan old, new
	}{
		{".env.example", []string{appNameKV, "APP_NAME=" + name}},
		{".env", []string{appNameKV, "APP_NAME=" + name}},
		{"internal/config/config.go", []string{`"APP_NAME", "` + templateName + `"`, `"APP_NAME", "` + name + `"`}},
		{"internal/handler/health_handler.go", []string{`example:"` + templateName + `"`, `example:"` + name + `"`}},
		{"cmd/api/main.go", []string{templateTitle, title}},
		// File Swagger hasil generate diperbarui langsung supaya tidak wajib
		// menjalankan swag; hasilnya sama dengan menjalankan `swag init` ulang.
		{"docs/docs.go", []string{templateTitle, title, templateName, name}},
		{"docs/swagger.json", []string{templateTitle, title, templateName, name}},
		{"docs/swagger.yaml", []string{templateTitle, title, templateName, name}},
		{"README.md", []string{"# Gin Template\n", "# " + title + "\n", "# Gin Template\r\n", "# " + title + "\r\n"}},
	}

	var changed []string
	for _, e := range edits {
		path := filepath.Join(root, filepath.FromSlash(e.path))
		info, err := os.Stat(path)
		if errors.Is(err, os.ErrNotExist) {
			continue // mis. .env belum dibuat
		}
		if err != nil {
			return changed, err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return changed, err
		}
		updated := strings.NewReplacer(e.replacements...).Replace(string(data))
		if updated == string(data) {
			continue
		}
		if err := os.WriteFile(path, []byte(updated), info.Mode().Perm()); err != nil {
			return changed, err
		}
		changed = append(changed, e.path)
	}
	return changed, nil
}

func readModulePath(goMod string) (string, error) {
	data, err := os.ReadFile(goMod)
	if err != nil {
		return "", fmt.Errorf("membaca go.mod: %w (jalankan dari root project)", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.Trim(strings.TrimSpace(rest), `"`), nil
		}
	}
	return "", errors.New("baris module tidak ditemukan di go.mod")
}

var majorVersion = regexp.MustCompile(`^v[0-9]+$`)

// nameFromModule: "github.com/user/shop-api" -> "shop-api",
// "github.com/user/shop-api/v2" -> "shop-api".
func nameFromModule(module string) string {
	parts := strings.Split(module, "/")
	if len(parts) > 1 && majorVersion.MatchString(parts[len(parts)-1]) {
		parts = parts[:len(parts)-1]
	}
	return parts[len(parts)-1]
}

// titleFromName: "shop-api" -> "Shop API", "nama_project_baru" -> "Nama Project Baru API".
func titleFromName(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool { return r == '-' || r == '_' || r == '.' || r == ' ' })
	if n := len(words); n > 0 && strings.EqualFold(words[n-1], "api") {
		words = words[:n-1]
	}
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.TrimSpace(strings.Join(words, " ") + " API")
}
