package demod

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"
)

type SyncOptions struct {
	DryRun bool
	Logger *slog.Logger
}

func (o SyncOptions) logger() *slog.Logger {
	if o.Logger != nil {
		return o.Logger
	}
	return slog.Default()
}

func SyncAll(cfg *Config, opts SyncOptions) error {
	g, ctx := errgroup.WithContext(context.Background())
	for _, mod := range cfg.Modules {
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return SyncModule(mod, opts)
			}
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	opts.logger().Info("done")
	return nil
}

func SyncModule(mod Module, opts SyncOptions) error {
	logger := opts.logger()

	tmpdir, err := os.MkdirTemp("", "demod-*")
	if err != nil {
		return fmt.Errorf("[%s] creating temp dir: %w", mod.Name, err)
	}
	defer func() { _ = os.RemoveAll(tmpdir) }()

	workdir := filepath.Join(tmpdir, "repo")

	logger.Info("cloning", "module", mod.Name)
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

	logger.Info("checkout", "module", mod.Name, "revision", mod.Revision)
	if err := gitCheckout(workdir, mod.Revision); err != nil {
		return fmt.Errorf("[%s] %w", mod.Name, err)
	}

	if opts.DryRun {
		logger.Info("would sync", "module", mod.Name, "dest", mod.Dest)
		for _, p := range mod.Paths {
			destPath := p.As
			if destPath == "" {
				destPath = p.Src
			}
			logger.Info("would copy", "module", mod.Name, "src", p.Src, "dest", filepath.Join(mod.Dest, destPath))
		}
		return nil
	}

	logger.Info("syncing", "module", mod.Name, "dest", mod.Dest)

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
