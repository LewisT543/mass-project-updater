package git

import (
	"os/exec"
)

type RealGit struct{}

func New() Git {
	return &RealGit{}
}

func run(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Run()
}

func (g *RealGit) Clone(url, dir string) error {
	return run("", "clone", url, dir)
}

func (g *RealGit) CheckoutDevelop(dir string) error {
	if err := run(dir, "checkout", "develop"); err != nil {
		return err
	}
	return run(dir, "pull")
}

func (g *RealGit) CreateBranch(dir, name string) error {
	return run(dir, "checkout", "-b", name)
}

func (g *RealGit) CommitAndPush(dir, msg, branch string) error {
	if err := run(dir, "add", "."); err != nil {
		return err
	}
	if err := run(dir, "commit", "-m", msg); err != nil {
		return err
	}
	return run(dir, "push", "-u", "origin", branch)
}
