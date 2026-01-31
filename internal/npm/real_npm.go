package npm

import "os/exec"

type RealNPM struct{}

func New() NPM {
	return &RealNPM{}
}

func run(dir string, args ...string) error {
	cmd := exec.Command("npm", args...)
	cmd.Dir = dir
	return cmd.Run()
}

func (n *RealNPM) Install(dir string) error {
	return run(dir, "install")
}

func (n *RealNPM) Build(dir string) error {
	return run(dir, "run", "build")
}
