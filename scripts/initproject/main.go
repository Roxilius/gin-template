// Command initproject membuat project baru dari template ini, atau
// menyesuaikan project hasil salinan template.
//
// Membuat project baru (cukup nama folder, tanpa clone):
//
//	go run github.com/Roxilius/gin-template/scripts/initproject@main shop-api
//
// Menyesuaikan folder yang sudah disalin manual (jalankan dari root-nya):
//
//	go run ./scripts/initproject
//
// Keduanya mengganti module path di go.mod dan semua import (default: nama
// folder), APP_NAME, judul Swagger, dan judul README, lalu membuat .env.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
)

const (
	templateName  = "gin-template"
	templateTitle = "Gin Template API"
	// Disusun dari dua bagian supaya tidak ikut terganti saat module di-rename.
	templateModule = "github.com/Roxilius/" + templateName
)

type options struct {
	module string // module path baru; kosong = nama folder
	name   string // APP_NAME; kosong = segmen terakhir module path
	title  string // judul Swagger; kosong = diturunkan dari name
}

func main() {
	var opt options
	var from string
	flag.StringVar(&opt.module, "module", "", `module path baru (default: nama folder, mis. "shop-api")`)
	flag.StringVar(&opt.name, "name", "", "APP_NAME (default: segmen terakhir module path)")
	flag.StringVar(&opt.title, "title", "", `judul Swagger (default: diturunkan dari nama, mis. "Shop API")`)
	flag.StringVar(&from, "from", "", "sumber template: folder lokal atau module@versi (default: versi yang sedang dijalankan)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Pemakaian:")
		fmt.Fprintln(os.Stderr, "  initproject [flag] <folder-baru>   buat project baru dari template")
		fmt.Fprintln(os.Stderr, "  initproject [flag]                 sesuaikan project di folder saat ini")
		flag.PrintDefaults()
	}
	flag.Parse()

	root := "."
	switch flag.NArg() {
	case 0:
	case 1:
		root = flag.Arg(0)
		if err := checkTarget(root, opt.module); err != nil {
			fail(err)
		}
		src, err := resolveSource(from)
		if err == nil {
			err = copyTree(src, root)
		}
		if err != nil {
			fail(err)
		}
		fmt.Printf("Template disalin dari %s ke %s\n", src, root)
	default:
		flag.Usage()
		os.Exit(2)
	}

	changed, err := run(root, opt)
	if err != nil {
		fail(err)
	}
	if len(changed) == 0 {
		fmt.Println("Tidak ada yang perlu diubah (mungkin sudah pernah dijalankan).")
		return
	}
	fmt.Println("File yang diubah:")
	for _, f := range changed {
		fmt.Println("  -", f)
	}
	if flag.NArg() == 1 {
		fmt.Printf("\nSelesai. Lanjutkan dengan:\n  cd %s\n  ./scripts/dev.ps1 db-up   (Linux/macOS: make db-up)\n  ./scripts/dev.ps1 run     (Linux/macOS: make run)\n", root)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

// run menyesuaikan project di root dan mengembalikan daftar file yang berubah.
func run(root string, opt options) ([]string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	current, err := readModulePath(filepath.Join(absRoot, "go.mod"))
	if err != nil {
		return nil, err
	}

	var changed []string

	// 1. Module path. Default-nya nama folder, hanya jika masih nama template.
	module := opt.module
	if module == "" && current == templateModule {
		module = filepath.Base(absRoot)
	}
	if module != "" && module != current {
		if nameFromModule(module) == templateName {
			return nil, errors.New("folder ini masih bernama " + templateName +
				" (repo template-nya sendiri?); buat project di folder baru, atau beri -module")
		}
		if !validModule.MatchString(module) {
			return nil, fmt.Errorf("module path %q tidak valid: gunakan huruf, angka, '-', '_', '.', dan '/'", module)
		}
		files, err := renameModule(absRoot, current, module)
		if err != nil {
			return files, err
		}
		changed = append(changed, fmt.Sprintf("go.mod + %d file .go (module %s -> %s)", len(files)-1, current, module))
		current = module
	}

	// 2. Nama aplikasi.
	name := opt.name
	if name == "" {
		name = nameFromModule(current)
	}
	if name == templateName {
		return nil, errors.New("nama project masih " + templateName +
			"; jalankan di project baru, atau beri -name")
	}
	title := opt.title
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
	for _, e := range edits {
		ok, err := replaceInFile(filepath.Join(absRoot, filepath.FromSlash(e.path)), strings.NewReplacer(e.replacements...))
		if err != nil {
			return changed, err
		}
		if ok && !contains(changed, e.path) {
			changed = append(changed, e.path)
		}
	}

	// 3. .env dari .env.example, jika belum ada.
	env := filepath.Join(absRoot, ".env")
	if _, err := os.Stat(env); errors.Is(err, os.ErrNotExist) {
		if data, err := os.ReadFile(filepath.Join(absRoot, ".env.example")); err == nil {
			if err := os.WriteFile(env, data, 0o600); err != nil {
				return changed, err
			}
			changed = append(changed, ".env (dibuat dari .env.example)")
		}
	}
	return changed, nil
}

// checkTarget memvalidasi folder tujuan sebelum apa pun disalin.
func checkTarget(dst, module string) error {
	abs, err := filepath.Abs(dst)
	if err != nil {
		return err
	}
	if module == "" {
		module = filepath.Base(abs)
	}
	if !validModule.MatchString(module) {
		return fmt.Errorf("nama %q tidak valid: gunakan huruf, angka, '-', '_', '.' (contoh: shop-api)", module)
	}
	if nameFromModule(module) == templateName {
		return errors.New("pilih nama project selain " + templateName)
	}
	if entries, err := os.ReadDir(abs); err == nil && len(entries) > 0 {
		return fmt.Errorf("folder %s sudah ada dan tidak kosong", dst)
	}
	return nil
}

var validModule = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._~-]*(/[A-Za-z0-9][A-Za-z0-9._~-]*)*$`)

// renameModule mengganti module path di go.mod dan semua import di file .go.
func renameModule(root, oldMod, newMod string) ([]string, error) {
	var changed []string
	goMod := strings.NewReplacer("module "+oldMod+"\n", "module "+newMod+"\n", "module "+oldMod+"\r\n", "module "+newMod+"\r\n")
	imports := strings.NewReplacer(`"`+oldMod+`"`, `"`+newMod+`"`, `"`+oldMod+`/`, `"`+newMod+`/`)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && skipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		var r *strings.Replacer
		switch {
		case path == filepath.Join(root, "go.mod"):
			r = goMod
		case strings.HasSuffix(path, ".go"):
			r = imports
		default:
			return nil
		}
		ok, err := replaceInFile(path, r)
		if ok {
			rel, _ := filepath.Rel(root, path)
			changed = append(changed, filepath.ToSlash(rel))
		}
		return err
	})
	return changed, err
}

func replaceInFile(path string, r *strings.Replacer) (bool, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	updated := r.Replace(string(data))
	if updated == string(data) {
		return false, nil
	}
	return true, os.WriteFile(path, []byte(updated), info.Mode().Perm())
}

// resolveSource menentukan folder sumber template.
func resolveSource(from string) (string, error) {
	if from == "" {
		if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Path == templateModule &&
			bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			// Dijalankan lewat `go run .../initproject@versi`: salin versi yang sama.
			from = templateModule + "@" + bi.Main.Version
		} else {
			// Dijalankan dari clone lokal: salin folder template saat ini.
			from = "."
		}
	}
	if strings.Contains(from, "@") {
		return downloadModule(from)
	}
	abs, err := filepath.Abs(from)
	if err != nil {
		return "", err
	}
	if mod, err := readModulePath(filepath.Join(abs, "go.mod")); err != nil || mod != templateModule {
		return "", fmt.Errorf("%s bukan folder template %s", abs, templateModule)
	}
	return abs, nil
}

// downloadModule mengunduh module ke module cache dan mengembalikan foldernya.
func downloadModule(modVersion string) (string, error) {
	cmd := exec.Command("go", "mod", "download", "-json", modVersion)
	cmd.Dir = os.TempDir() // di luar module mana pun
	out, err := cmd.Output()
	var info struct{ Dir, Error string }
	_ = json.Unmarshal(out, &info)
	if info.Error != "" {
		return "", fmt.Errorf("mengunduh %s: %s", modVersion, info.Error)
	}
	if err != nil {
		return "", fmt.Errorf("mengunduh %s: %w", modVersion, err)
	}
	if info.Dir == "" {
		return "", fmt.Errorf("mengunduh %s: folder module tidak diketahui", modVersion)
	}
	return info.Dir, nil
}

// copyTree menyalin template ke dst. dst harus belum ada atau kosong.
func copyTree(src, dst string) error {
	absDst, err := filepath.Abs(dst)
	if err != nil {
		return err
	}
	if entries, err := os.ReadDir(absDst); err == nil && len(entries) > 0 {
		return fmt.Errorf("folder %s sudah ada dan tidak kosong", dst)
	}

	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(absDst, rel)
		if d.IsDir() {
			if path == absDst || (rel != "." && skipDir(d.Name())) {
				return filepath.SkipDir
			}
			return os.MkdirAll(target, 0o755)
		}
		if skipFile(d.Name()) {
			return nil
		}
		// File di module cache bersifat read-only; salinannya dibuat bisa diedit.
		return copyFile(path, target, 0o644)
	})
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// File dan folder milik lingkungan lokal yang tidak ikut disalin.
func skipDir(name string) bool {
	switch name {
	case ".git", "tmp", "bin", "node_modules", ".idea", ".vscode":
		return true
	}
	return false
}

func skipFile(name string) bool {
	switch name {
	case ".env", "coverage.out", "coverage.html":
		return true
	}
	return false
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
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
// "github.com/user/shop-api/v2" -> "shop-api", "shop-api" -> "shop-api".
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
