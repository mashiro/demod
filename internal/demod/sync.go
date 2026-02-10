package demod

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func SyncAll(cfg *Config) error {
	for _, mod := range cfg.Modules {
		if err := SyncModule(mod); err != nil {
			return err
		}
	}
	fmt.Println("Done.")
	return nil
}

func SyncModule(mod Module) error {
	tmpdir, err := os.MkdirTemp("", "demod-*")
	if err != nil {
		return fmt.Errorf("[%s] creating temp dir: %w", mod.Name, err)
	}
	defer func() { _ = os.RemoveAll(tmpdir) }()

	workdir := filepath.Join(tmpdir, "repo")

	fmt.Printf("[%s] cloning...\n", mod.Name)
	if err := gitClone(mod.Repo, workdir); err != nil {
		return fmt.Errorf("[%s] %w", mod.Name, err)
	}

	if err := gitSparseCheckoutInit(workdir); err != nil {
		return fmt.Errorf("[%s] %w", mod.Name, err)
	}

	if err := gitSparseCheckoutSet(workdir, mod.Paths); err != nil {
		return fmt.Errorf("[%s] %w", mod.Name, err)
	}

	fmt.Printf("[%s] checkout %s\n", mod.Name, mod.Revision)
	if err := gitCheckout(workdir, mod.Revision); err != nil {
		return fmt.Errorf("[%s] %w", mod.Name, err)
	}

	fmt.Printf("[%s] syncing to %s\n", mod.Name, mod.Dest)

	if err := os.RemoveAll(mod.Dest); err != nil {
		return fmt.Errorf("[%s] removing dest: %w", mod.Name, err)
	}

	for _, p := range mod.Paths {
		src := filepath.Join(workdir, p)
		if err := copyDir(src, mod.Dest, p, mod.StripPrefix); err != nil {
			return fmt.Errorf("[%s] copying: %w", mod.Name, err)
		}
	}

	return nil
}

func copyDir(src, dest, path, stripPrefix string) error {
	return filepath.Walk(src, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, fpath)
		if err != nil {
			return err
		}

		// Build destination path: dest + path (with stripPrefix applied) + rel
		destPath := path
		if stripPrefix != "" {
			destPath = strings.TrimPrefix(destPath, stripPrefix)
			destPath = strings.TrimPrefix(destPath, string(filepath.Separator))
		}

		target := filepath.Join(dest, destPath, rel)

		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		return copyFile(fpath, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(dst)
		return err
	}

	return out.Close()
}
