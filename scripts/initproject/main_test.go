package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNameAndTitle(t *testing.T) {
	cases := []struct{ module, name, title string }{
		{"shop-api", "shop-api", "Shop API"},
		{"github.com/user/shop-api", "shop-api", "Shop API"},
		{"github.com/user/nama_project_baru", "nama_project_baru", "Nama Project Baru API"},
		{"github.com/user/inventory/v2", "inventory", "Inventory API"},
		{"example.com/api", "api", "API"},
	}
	for _, c := range cases {
		if got := nameFromModule(c.module); got != c.name {
			t.Errorf("nameFromModule(%q) = %q, want %q", c.module, got, c.name)
		}
		if got := titleFromName(c.name); got != c.title {
			t.Errorf("titleFromName(%q) = %q, want %q", c.name, got, c.title)
		}
	}
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func exists(root, rel string) bool {
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	return err == nil
}

// writeTemplate membuat versi mini template di root.
func writeTemplate(t *testing.T, root string) {
	t.Helper()
	writeFile(t, root, "go.mod", "module "+templateModule+"\r\n\r\ngo 1.27\r\n")
	writeFile(t, root, ".env.example", "APP_NAME=gin-template\nAPP_PORT=8080\n")
	writeFile(t, root, "cmd/api/main.go", "//\t@title\t\tGin Template API\npackage main\n\nimport (\n\t_ \""+
		templateModule+"/docs\"\n\t\""+templateModule+"/internal/config\"\n)\n")
	writeFile(t, root, "README.md", "# Gin Template\r\n\r\nisi\r\n")
	writeFile(t, root, "docs/swagger.json", `{"title": "Gin Template API", "example": "gin-template"}`)
}

func TestRunRenamesModuleFromFolderName(t *testing.T) {
	root := filepath.Join(t.TempDir(), "shop-api")
	writeTemplate(t, root)

	if _, err := run(root, options{}); err != nil {
		t.Fatal(err)
	}

	checks := map[string]string{
		"go.mod":            "module shop-api\r\n",
		"cmd/api/main.go":   `"shop-api/internal/config"`,
		".env.example":      "APP_NAME=shop-api\n",
		".env":              "APP_NAME=shop-api\n",
		"README.md":         "# Shop API\r\n",
		"docs/swagger.json": `{"title": "Shop API", "example": "shop-api"}`,
	}
	for file, want := range checks {
		if got := readFile(t, root, file); !strings.Contains(got, want) {
			t.Errorf("%s = %q, want to contain %q", file, got, want)
		}
	}
	if got := readFile(t, root, "cmd/api/main.go"); strings.Contains(got, templateModule) || !strings.Contains(got, `_ "shop-api/docs"`) {
		t.Errorf("imports not fully renamed:\n%s", got)
	}

	// Idempotent: dijalankan ulang tidak mengubah apa pun.
	if changed, err := run(root, options{}); err != nil || len(changed) != 0 {
		t.Errorf("second run: changed=%v err=%v, want none", changed, err)
	}
}

func TestRunWithExplicitModule(t *testing.T) {
	root := filepath.Join(t.TempDir(), "folder-apa-saja")
	writeTemplate(t, root)

	if _, err := run(root, options{module: "github.com/user/toko"}); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, root, "go.mod"); !strings.Contains(got, "module github.com/user/toko\r\n") {
		t.Errorf("go.mod = %q", got)
	}
	if got := readFile(t, root, ".env"); !strings.Contains(got, "APP_NAME=toko") {
		t.Errorf(".env = %q", got)
	}
}

func TestRunRefusesTemplateItself(t *testing.T) {
	root := filepath.Join(t.TempDir(), templateName)
	writeTemplate(t, root)
	if _, err := run(root, options{}); err == nil {
		t.Fatal("expected error when run inside the template itself")
	}
	if got := readFile(t, root, "go.mod"); !strings.Contains(got, templateModule) {
		t.Errorf("template go.mod was modified: %q", got)
	}
}

func TestCopyTree(t *testing.T) {
	src := t.TempDir()
	writeTemplate(t, src)
	writeFile(t, src, ".env", "DB_PASSWORD=rahasia\n")
	writeFile(t, src, ".git/HEAD", "ref: refs/heads/main\n")
	writeFile(t, src, "tmp/main.exe", "bin")
	writeFile(t, src, ".github/workflows/ci.yml", "name: CI\n")

	dst := filepath.Join(t.TempDir(), "shop-api")
	if err := copyTree(src, dst); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"go.mod", "cmd/api/main.go", ".env.example", ".github/workflows/ci.yml"} {
		if !exists(dst, f) {
			t.Errorf("%s not copied", f)
		}
	}
	for _, f := range []string{".env", ".git", "tmp"} {
		if exists(dst, f) {
			t.Errorf("%s should not be copied", f)
		}
	}

	if err := copyTree(src, dst); err == nil {
		t.Error("expected error when destination is not empty")
	}
}

func TestCheckTarget(t *testing.T) {
	base := t.TempDir()
	if err := checkTarget(filepath.Join(base, "shop-api"), ""); err != nil {
		t.Errorf("valid name rejected: %v", err)
	}
	for _, bad := range []string{"nama project", templateName} {
		if err := checkTarget(filepath.Join(base, bad), ""); err == nil {
			t.Errorf("checkTarget(%q) = nil, want error", bad)
		}
	}
}
