package demod

import (
	"fmt"
	"log/slog"
	"os/exec"
)

func runGit(logger *slog.Logger, workdir string, args ...string) error {
	logger.Debug("exec", "cmd", "git", "args", args)
	cmd := exec.Command("git", args...)
	cmd.Dir = workdir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %w\n%s", args[0], err, out)
	}
	logger.Debug("output", "result", string(out))
	return nil
}

func gitClone(logger *slog.Logger, repo, workdir string) error {
	return runGit(logger, "", "clone", "--filter=blob:none", "--no-checkout", "--depth", "1", repo, workdir)
}

func gitSparseCheckoutInit(logger *slog.Logger, workdir string) error {
	return runGit(logger, workdir, "sparse-checkout", "init", "--cone")
}

func gitSparseCheckoutSet(logger *slog.Logger, workdir string, paths []string) error {
	args := append([]string{"sparse-checkout", "set"}, paths...)
	return runGit(logger, workdir, args...)
}

func gitCheckout(logger *slog.Logger, workdir, revision string) error {
	return runGit(logger, workdir, "checkout", revision)
}
