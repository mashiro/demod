package demod

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRunGit(t *testing.T) {
	logger := slog.Default()

	t.Run("success", func(t *testing.T) {
		dir := t.TempDir()
		if err := exec.Command("git", "init", dir).Run(); err != nil {
			t.Fatal(err)
		}
		if err := runGit(logger, dir, "status"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		dir := t.TempDir()
		if err := runGit(logger, dir, "checkout", "nonexistent"); err == nil {
			t.Fatal("expected error for invalid git command")
		}
	})
}

func TestGitClone(t *testing.T) {
	logger := slog.Default()
	bare := setupBareRepo(t)
	workdir := filepath.Join(t.TempDir(), "repo")

	if err := gitClone(logger, bare, workdir); err != nil {
		t.Fatalf("gitClone: %v", err)
	}

	if _, err := os.Stat(filepath.Join(workdir, ".git")); err != nil {
		t.Fatalf("expected .git dir in cloned repo: %v", err)
	}
}

func TestGitSparseCheckout(t *testing.T) {
	logger := slog.Default()
	bare := setupBareRepo(t)
	workdir := filepath.Join(t.TempDir(), "repo")

	if err := gitClone(logger, bare, workdir); err != nil {
		t.Fatalf("gitClone: %v", err)
	}

	if err := gitSparseCheckoutInit(logger, workdir); err != nil {
		t.Fatalf("gitSparseCheckoutInit: %v", err)
	}

	if err := gitSparseCheckoutSet(logger, workdir, []string{"src/lib"}); err != nil {
		t.Fatalf("gitSparseCheckoutSet: %v", err)
	}

	if err := gitCheckout(logger, workdir, "main"); err != nil {
		t.Fatalf("gitCheckout: %v", err)
	}

	// src/lib/a.txt should exist
	if _, err := os.Stat(filepath.Join(workdir, "src", "lib", "a.txt")); err != nil {
		t.Errorf("expected src/lib/a.txt to exist: %v", err)
	}

	// docs/readme.txt should NOT exist (sparse checkout)
	if _, err := os.Stat(filepath.Join(workdir, "docs", "readme.txt")); !os.IsNotExist(err) {
		t.Errorf("expected docs/readme.txt to not exist, got err: %v", err)
	}
}

// setupBareRepo creates a bare git repo with the following structure:
//
//	src/lib/a.txt  ("aaa")
//	src/lib/b.txt  ("bbb")
//	docs/readme.txt ("readme")
func setupBareRepo(t *testing.T) string {
	t.Helper()

	// Create a normal repo and populate it
	workdir := filepath.Join(t.TempDir(), "work")
	if err := exec.Command("git", "init", workdir).Run(); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = workdir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")

	for _, f := range []struct {
		path, content string
	}{
		{"src/lib/a.txt", "aaa"},
		{"src/lib/b.txt", "bbb"},
		{"docs/readme.txt", "readme"},
	} {
		abs := filepath.Join(workdir, f.path)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(f.content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	run("add", "-A")
	run("commit", "-m", "initial")
	run("branch", "-M", "main")

	// Create a bare clone
	bare := filepath.Join(t.TempDir(), "bare.git")
	if out, err := exec.Command("git", "clone", "--bare", workdir, bare).CombinedOutput(); err != nil {
		t.Fatalf("creating bare repo: %v\n%s", err, out)
	}

	return bare
}
