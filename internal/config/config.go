package config

import "os"

type Config struct {
	GitlabToken string
	GitlabBase  string
	GroupID     string
	WorkDir     string
	BranchName  string
}

func Load() Config {
	return Config{
		GitlabToken: os.Getenv("GITLAB_TOKEN"),
		GitlabBase:  os.Getenv("GITLAB_BASE_URL"),
		GroupID:     os.Getenv("GITLAB_GROUP_ID"),
		WorkDir:     os.Getenv("WORKDIR"),
		BranchName:  getenv("BRANCH_NAME", "chore/dependency-updates"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
