package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNameAndTitle(t *testing.T) {
	cases := []struct{ module, name, title string }{
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

func TestRun(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module github.com/user/shop-api\n\ngo 1.27\n")
	writeFile(t, root, ".env.example", "APP_NAME=gin-template\nAPP_PORT=8080\n")
	writeFile(t, root, "cmd/api/main.go", "//\t@title\t\tGin Template API\npackage main\n")
	writeFile(t, root, "README.md", "# Gin Template\r\n\r\nisi\r\n")
	writeFile(t, root, "docs/swagger.json", `{"title": "Gin Template API", "example": "gin-template"}`)

	changed, err := run(root, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(changed) != 4 {
		t.Errorf("changed = %v, want 4 files", changed)
	}

	checks := map[string]string{
		".env.example":      "APP_NAME=shop-api\n",
		"cmd/api/main.go":   "Shop API",
		"README.md":         "# Shop API\r\n",
		"docs/swagger.json": `{"title": "Shop API", "example": "shop-api"}`,
	}
	for file, want := range checks {
		if got := readFile(t, root, file); !strings.Contains(got, want) {
			t.Errorf("%s = %q, want to contain %q", file, got, want)
		}
	}

	// Idempotent: dijalankan ulang tidak mengubah apa pun.
	if changed, err := run(root, "", ""); err != nil || len(changed) != 0 {
		t.Errorf("second run: changed=%v err=%v, want none", changed, err)
	}
}

func TestRunRefusesTemplateItself(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module github.com/someone/"+templateName+"\n")
	if _, err := run(root, "", ""); err == nil {
		t.Fatal("expected error when run inside the template itself")
	}
}
