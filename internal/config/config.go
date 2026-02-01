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
	BranchName  string
	MaxWorkers  int
	DryRun      bool
}

// TODO Set up logging dir in this?
func Load() Config {
	cwd, _ := os.Getwd()
	workDir := filepath.Join(cwd, "..", "UI_PROJECTS_FOR_BOT")
	os.MkdirAll(workDir, 0755)

	return Config{
		GitlabToken: os.Getenv("GITLAB_TOKEN"),
		GitlabBase:  os.Getenv("GITLAB_BASE_URL"),
		GroupID:     os.Getenv("GITLAB_GROUP_ID"),
		WorkDir:     workDir,
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
