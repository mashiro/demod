package demod

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
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
	logger := WithModule(opts.logger(), mod.Name)

	tmpdir, err := os.MkdirTemp("", "demod-*")
	if err != nil {
		return fmt.Errorf("[%s] creating temp dir: %w", mod.Name, err)
	}
	defer func() { _ = os.RemoveAll(tmpdir) }()

	workdir := filepath.Join(tmpdir, "repo")

	logger.Info("cloning")
	if err := gitClone(logger, mod.Repo, workdir); err != nil {
		return fmt.Errorf("[%s] %w", mod.Name, err)
	}

	if err := gitSparseCheckoutInit(logger, workdir); err != nil {
		return fmt.Errorf("[%s] %w", mod.Name, err)
	}

	srcPaths := make([]string, len(mod.Paths))
	for i, p := range mod.Paths {
		srcPaths[i] = p.Src
	}
	if err := gitSparseCheckoutSet(logger, workdir, srcPaths); err != nil {
		return fmt.Errorf("[%s] %w", mod.Name, err)
	}

	logger.Info("checkout", "revision", mod.Revision)
	if err := gitCheckout(logger, workdir, mod.Revision); err != nil {
		return fmt.Errorf("[%s] %w", mod.Name, err)
	}

	if opts.DryRun {
		logger.Info("would sync", "dest", mod.Dest)
		for _, p := range mod.Paths {
			destPath := p.As
			if destPath == "" {
				destPath = p.Src
			}
			logger.Info("would copy", "src", p.Src, "dest", filepath.Join(mod.Dest, destPath), "exclude", p.Exclude)
		}
		return nil
	}

	logger.Info("syncing", "dest", mod.Dest)

	if err := os.RemoveAll(mod.Dest); err != nil {
		return fmt.Errorf("[%s] removing dest: %w", mod.Name, err)
	}

	for _, p := range mod.Paths {
		src := filepath.Join(workdir, p.Src)
		destPath := p.As
		if destPath == "" {
			destPath = p.Src
		}
		if err := copyDir(src, mod.Dest, destPath, p.Exclude); err != nil {
			return fmt.Errorf("[%s] copying: %w", mod.Name, err)
		}
	}

	return nil
}

func copyDir(src, dest, destPath string, exclude []string) error {
	return filepath.WalkDir(src, func(fpath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, fpath)
		if err != nil {
			return err
		}

		if rel != "." {
			for _, pattern := range exclude {
				matched, matchErr := doublestar.Match(pattern, rel)
				if matchErr != nil {
					return fmt.Errorf("invalid exclude pattern %q: %w", pattern, matchErr)
				}
				if matched {
					if d.IsDir() {
						return fs.SkipDir
					}
					return nil
				}
			}
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
