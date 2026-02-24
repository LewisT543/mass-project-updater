//go:build ignore

package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	distDir := filepath.FromSlash("./dist")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		log.Fatal(err)
	}

	exePath := filepath.Join(distDir, "mass-project-updater.exe")
	cmd := exec.Command("go", "build", "-o", exePath, "./cmd/updater")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	ensureDepsJSON(distDir)

	// Optionally include the README for distribution (ignore if missing).
	if readme, err := os.ReadFile(filepath.FromSlash("./README.md")); err == nil {
		_ = os.WriteFile(filepath.Join(distDir, "README.md"), readme, 0644)
	}
}

// ensureDepsJSON creates a starter deps.json in distDir if one does not already exist.
func ensureDepsJSON(distDir string) {
	depsPath := filepath.Join(distDir, "deps.json")
	if _, err := os.Stat(depsPath); !os.IsNotExist(err) {
		return
	}

	starter := map[string]interface{}{
		"dependencies":    map[string]string{},
		"devDependencies": map[string]string{},
		"overrides":       map[string]interface{}{},
	}

	b, err := json.MarshalIndent(starter, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(depsPath, append(b, '\n'), 0644); err != nil {
		log.Fatal(err)
	}
}