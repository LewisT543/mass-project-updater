package main

import (
	"fmt"
	"go.uber.org/zap"
	"lewist543.com/mass-project-updater/internal/config"
	"lewist543.com/mass-project-updater/internal/git"
	"lewist543.com/mass-project-updater/internal/gitlab"
	"lewist543.com/mass-project-updater/internal/model"
	"lewist543.com/mass-project-updater/internal/npm"
	"lewist543.com/mass-project-updater/internal/packagejson"
	"lewist543.com/mass-project-updater/internal/runner"
	"log"
	"os"
)

func main() {
	cfg := config.Load()
	if err := prepareFilesystem(cfg); err != nil {
		log.Fatal(err)
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	updates, err := packagejson.LoadUpdates("deps.json")
	if err != nil {
		log.Fatal(err)
	}

	projects, err := gitlab.FetchUIProjects(
		cfg.GitlabBase,
		cfg.GitlabToken,
		cfg.GroupID,
		"ui-spa",
	)
	if err != nil {
		log.Fatal(err)
	}

	gitRunner := git.New()
	npmRunner := npm.New()
	mrClient := gitlab.NewMRClient(cfg.GitlabBase, cfg.GitlabToken)
	maxWorkers := 5

	results := runner.RunAllProjects(projects, maxWorkers, func(p model.Project) (string, error) {
		mrURL, err := runner.RunProject(cfg, p, updates, gitRunner, npmRunner, mrClient, "", logger)
		if err != nil {
			log.Printf("[ERROR] Project %s: %v", p.Name, err)
		} else {
			log.Printf("[OK] Project %s processed, MR: %s", p.Name, mrURL)
		}
		return mrURL, err
	})

	// Report summary
	var failed int
	for _, r := range results {
		if r.Err != nil {
			failed++
			log.Printf("[FAIL] %s: %v", r.Project.Name, r.Err)
		} else {
			log.Printf("[SUCCESS] %s MR: %s", r.Project.Name, r.MRURL)
		}
	}

	fmt.Printf("Completed %d projects: %d failed, %d succeeded\n", len(projects), failed, len(projects)-failed)
}

func prepareFilesystem(cfg config.Config) error {
	if err := os.MkdirAll(cfg.WorkDir, 0755); err != nil {
		return err
	}

	if err := os.RemoveAll(cfg.ErrorLogDir); err != nil {
		return err
	}

	return os.MkdirAll(cfg.ErrorLogDir, 0755)
}
