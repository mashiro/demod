package demod

import (
	"fmt"
	"os/exec"
)

func runGit(workdir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = workdir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %w\n%s", args[0], err, out)
	}
	return nil
}

func gitClone(repo, workdir string) error {
	cmd := exec.Command("git", "clone", "--filter=blob:none", "--no-checkout", "--depth", "1", repo, workdir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone: %w\n%s", err, out)
	}
	return nil
}

func gitSparseCheckoutInit(workdir string) error {
	return runGit(workdir, "sparse-checkout", "init", "--cone")
}

func gitSparseCheckoutSet(workdir string, paths []string) error {
	args := append([]string{"sparse-checkout", "set"}, paths...)
	return runGit(workdir, args...)
}

func gitCheckout(workdir, revision string) error {
	return runGit(workdir, "checkout", revision)
}
