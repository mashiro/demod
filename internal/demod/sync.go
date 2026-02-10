package demod

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"
)

func SyncAll(cfg *Config, dryRun bool) error {
	g, ctx := errgroup.WithContext(context.Background())
	for _, mod := range cfg.Modules {
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return SyncModule(mod, dryRun)
			}
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	fmt.Println("Done.")
	return nil
}

func SyncModule(mod Module, dryRun bool) error {
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

	srcPaths := make([]string, len(mod.Paths))
	for i, p := range mod.Paths {
		srcPaths[i] = p.Src
	}
	if err := gitSparseCheckoutSet(workdir, srcPaths); err != nil {
		return fmt.Errorf("[%s] %w", mod.Name, err)
	}

	fmt.Printf("[%s] checkout %s\n", mod.Name, mod.Revision)
	if err := gitCheckout(workdir, mod.Revision); err != nil {
		return fmt.Errorf("[%s] %w", mod.Name, err)
	}

	if dryRun {
		fmt.Printf("[%s] would sync to %s\n", mod.Name, mod.Dest)
		for _, p := range mod.Paths {
			destPath := p.As
			if destPath == "" {
				destPath = p.Src
			}
			fmt.Printf("[%s]   %s -> %s\n", mod.Name, p.Src, filepath.Join(mod.Dest, destPath))
		}
		return nil
	}

	fmt.Printf("[%s] syncing to %s\n", mod.Name, mod.Dest)

	if err := os.RemoveAll(mod.Dest); err != nil {
		return fmt.Errorf("[%s] removing dest: %w", mod.Name, err)
	}

	for _, p := range mod.Paths {
		src := filepath.Join(workdir, p.Src)
		destPath := p.As
		if destPath == "" {
			destPath = p.Src
		}
		if err := copyDir(src, mod.Dest, destPath); err != nil {
			return fmt.Errorf("[%s] copying: %w", mod.Name, err)
		}
	}

	return nil
}

func copyDir(src, dest, destPath string) error {
	return filepath.WalkDir(src, func(fpath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, fpath)
		if err != nil {
			return err
		}

		target := filepath.Join(dest, destPath, rel)

		if d.IsDir() {
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
