package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	GitlabToken string
	GitlabBase  string
	GroupID     string
	WorkDir     string
	ErrorLogDir string
	BranchName  string
	MaxWorkers  int
	DryRun      bool
}

func Load() Config {
	cwd, _ := os.Getwd()

	return Config{
		GitlabToken: os.Getenv("GITLAB_TOKEN"),
		GitlabBase:  os.Getenv("GITLAB_BASE_URL"),
		GroupID:     os.Getenv("GITLAB_GROUP_ID"),
		WorkDir:     filepath.Join(cwd, "..", "UI_PROJECTS"),
		ErrorLogDir: filepath.Join(cwd, "..", "ERROR_LOGS"),
		BranchName:  getenv("BRANCH_NAME", "chore/dependency-updates"),
		MaxWorkers:  5,
		DryRun:      os.Getenv("DRY_RUN") == "true",
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
