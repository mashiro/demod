package demod

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()

	src := filepath.Join(dir, "src.txt")
	if err := os.WriteFile(src, []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(dir, "sub", "dst.txt")
	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("reading dst: %v", err)
	}
	if string(got) != "hello world" {
		t.Errorf("content = %q, want %q", got, "hello world")
	}
}

func TestCopyDir(t *testing.T) {
	t.Run("with destPath", func(t *testing.T) {
		srcDir := t.TempDir()
		destDir := t.TempDir()

		// Create source structure: srcDir/a.txt, srcDir/sub/b.txt
		if err := os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("aaa"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(srcDir, "sub"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(srcDir, "sub", "b.txt"), []byte("bbb"), 0o644); err != nil {
			t.Fatal(err)
		}

		if err := copyDir(srcDir, destDir, "lib"); err != nil {
			t.Fatalf("copyDir: %v", err)
		}

		// Files should be at destDir/lib/a.txt, destDir/lib/sub/b.txt
		assertFileContent(t, filepath.Join(destDir, "lib", "a.txt"), "aaa")
		assertFileContent(t, filepath.Join(destDir, "lib", "sub", "b.txt"), "bbb")
	})

	t.Run("with nested destPath", func(t *testing.T) {
		srcDir := t.TempDir()
		destDir := t.TempDir()

		if err := os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("aaa"), 0o644); err != nil {
			t.Fatal(err)
		}

		// destPath="lib" → files at destDir/lib/a.txt
		if err := copyDir(srcDir, destDir, "lib"); err != nil {
			t.Fatalf("copyDir: %v", err)
		}

		assertFileContent(t, filepath.Join(destDir, "lib", "a.txt"), "aaa")
	})

	t.Run("empty destPath copies to root", func(t *testing.T) {
		srcDir := t.TempDir()
		destDir := t.TempDir()

		if err := os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("aaa"), 0o644); err != nil {
			t.Fatal(err)
		}

		// destPath="" → files at destDir/a.txt
		if err := copyDir(srcDir, destDir, ""); err != nil {
			t.Fatalf("copyDir: %v", err)
		}

		assertFileContent(t, filepath.Join(destDir, "a.txt"), "aaa")
	})
}

func TestSyncModule(t *testing.T) {
	bare := setupBareRepo(t)

	t.Run("with as", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "dest")
		mod := Module{
			Name:     "test",
			Repo:     bare,
			Revision: "main",
			Dest:     dest,
			Paths:    []Path{{Src: "src/lib", As: "lib"}},
		}
		if err := SyncModule(mod, false); err != nil {
			t.Fatalf("SyncModule: %v", err)
		}
		assertFileContent(t, filepath.Join(dest, "lib", "a.txt"), "aaa")
		assertFileContent(t, filepath.Join(dest, "lib", "b.txt"), "bbb")
	})

	t.Run("without as", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "dest")
		mod := Module{
			Name:     "test",
			Repo:     bare,
			Revision: "main",
			Dest:     dest,
			Paths:    []Path{{Src: "docs"}},
		}
		if err := SyncModule(mod, false); err != nil {
			t.Fatalf("SyncModule: %v", err)
		}
		assertFileContent(t, filepath.Join(dest, "docs", "readme.txt"), "readme")
	})

	t.Run("dry-run does not write files", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "dest")
		mod := Module{
			Name:     "test",
			Repo:     bare,
			Revision: "main",
			Dest:     dest,
			Paths:    []Path{{Src: "src/lib", As: "lib"}},
		}
		if err := SyncModule(mod, true); err != nil {
			t.Fatalf("SyncModule dry-run: %v", err)
		}
		if _, err := os.Stat(dest); !os.IsNotExist(err) {
			t.Errorf("expected dest dir to not exist in dry-run, got err: %v", err)
		}
	})

	t.Run("multiple paths", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "dest")
		mod := Module{
			Name:     "test",
			Repo:     bare,
			Revision: "main",
			Dest:     dest,
			Paths: []Path{
				{Src: "src/lib", As: "lib"},
				{Src: "docs"},
			},
		}
		if err := SyncModule(mod, false); err != nil {
			t.Fatalf("SyncModule: %v", err)
		}
		assertFileContent(t, filepath.Join(dest, "lib", "a.txt"), "aaa")
		assertFileContent(t, filepath.Join(dest, "docs", "readme.txt"), "readme")
	})
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	if string(got) != want {
		t.Errorf("%s: content = %q, want %q", path, got, want)
	}
}
