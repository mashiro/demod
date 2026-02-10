package demod

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("valid config with as", func(t *testing.T) {
		content := `
version = 1

[[modules]]
name = "foo"
repo = "https://github.com/example/foo"
revision = "main"
dest = "vendor/foo"
paths = [
  { src = "src/lib", as = "lib" },
]
`
		path := writeTempConfig(t, content)
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Version != 1 {
			t.Errorf("version = %d, want 1", cfg.Version)
		}
		if len(cfg.Modules) != 1 {
			t.Fatalf("len(modules) = %d, want 1", len(cfg.Modules))
		}
		mod := cfg.Modules[0]
		if mod.Name != "foo" {
			t.Errorf("name = %q, want %q", mod.Name, "foo")
		}
		if mod.Paths[0].Src != "src/lib" {
			t.Errorf("paths[0].src = %q, want %q", mod.Paths[0].Src, "src/lib")
		}
		if mod.Paths[0].As != "lib" {
			t.Errorf("paths[0].as = %q, want %q", mod.Paths[0].As, "lib")
		}
	})

	t.Run("valid config without as", func(t *testing.T) {
		content := `
version = 1

[[modules]]
name = "bar"
repo = "https://github.com/example/bar"
revision = "main"
dest = "vendor/bar"
paths = [
  { src = "lib" },
]
`
		path := writeTempConfig(t, content)
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		mod := cfg.Modules[0]
		if mod.Paths[0].Src != "lib" {
			t.Errorf("paths[0].src = %q, want %q", mod.Paths[0].Src, "lib")
		}
		if mod.Paths[0].As != "" {
			t.Errorf("paths[0].as = %q, want %q", mod.Paths[0].As, "")
		}
	})

	t.Run("missing src in paths", func(t *testing.T) {
		content := `
version = 1

[[modules]]
name = "foo"
repo = "https://github.com/example/foo"
revision = "main"
dest = "vendor/foo"
paths = [
  { as = "lib" },
]
`
		path := writeTempConfig(t, content)
		_, err := Load(path)
		if err == nil {
			t.Fatal("expected error for missing src in paths")
		}
	})

	t.Run("unsupported version", func(t *testing.T) {
		content := `
version = 2

[[modules]]
name = "foo"
repo = "https://github.com/example/foo"
revision = "main"
dest = "vendor/foo"
paths = [{ src = "src" }]
`
		path := writeTempConfig(t, content)
		_, err := Load(path)
		if err == nil {
			t.Fatal("expected error for unsupported version")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		content := `
version = 1

[[modules]]
repo = "https://github.com/example/foo"
revision = "main"
dest = "vendor/foo"
paths = [{ src = "src" }]
`
		path := writeTempConfig(t, content)
		_, err := Load(path)
		if err == nil {
			t.Fatal("expected error for missing name")
		}
	})

	t.Run("missing repo", func(t *testing.T) {
		content := `
version = 1

[[modules]]
name = "foo"
revision = "main"
dest = "vendor/foo"
paths = [{ src = "src" }]
`
		path := writeTempConfig(t, content)
		_, err := Load(path)
		if err == nil {
			t.Fatal("expected error for missing repo")
		}
	})

	t.Run("missing revision", func(t *testing.T) {
		content := `
version = 1

[[modules]]
name = "foo"
repo = "https://github.com/example/foo"
dest = "vendor/foo"
paths = [{ src = "src" }]
`
		path := writeTempConfig(t, content)
		_, err := Load(path)
		if err == nil {
			t.Fatal("expected error for missing revision")
		}
	})

	t.Run("missing dest", func(t *testing.T) {
		content := `
version = 1

[[modules]]
name = "foo"
repo = "https://github.com/example/foo"
revision = "main"
paths = [{ src = "src" }]
`
		path := writeTempConfig(t, content)
		_, err := Load(path)
		if err == nil {
			t.Fatal("expected error for missing dest")
		}
	})

	t.Run("missing paths", func(t *testing.T) {
		content := `
version = 1

[[modules]]
name = "foo"
repo = "https://github.com/example/foo"
revision = "main"
dest = "vendor/foo"
`
		path := writeTempConfig(t, content)
		_, err := Load(path)
		if err == nil {
			t.Fatal("expected error for missing paths")
		}
	})

	t.Run("duplicate dest path", func(t *testing.T) {
		content := `
version = 1

[[modules]]
name = "foo"
repo = "https://github.com/example/foo"
revision = "main"
dest = "vendor/foo"
paths = [
  { src = "src/a", as = "lib" },
  { src = "src/b", as = "lib" },
]
`
		path := writeTempConfig(t, content)
		_, err := Load(path)
		if err == nil {
			t.Fatal("expected error for duplicate dest path")
		}
	})

	t.Run("path traversal in as", func(t *testing.T) {
		content := `
version = 1

[[modules]]
name = "foo"
repo = "https://github.com/example/foo"
revision = "main"
dest = "vendor/foo"
paths = [
  { src = "src/lib", as = "../escape" },
]
`
		path := writeTempConfig(t, content)
		_, err := Load(path)
		if err == nil {
			t.Fatal("expected error for path traversal in as")
		}
	})

	t.Run("path traversal in src used as dest", func(t *testing.T) {
		content := `
version = 1

[[modules]]
name = "foo"
repo = "https://github.com/example/foo"
revision = "main"
dest = "vendor/foo"
paths = [
  { src = "../escape" },
]
`
		path := writeTempConfig(t, content)
		_, err := Load(path)
		if err == nil {
			t.Fatal("expected error for path traversal in src used as dest")
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := Load("/nonexistent/path/demod.toml")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "demod.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
