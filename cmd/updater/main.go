package main

import (
	"lewist543.com/mass-project-updater/internal/git"
	"lewist543.com/mass-project-updater/internal/npm"
	"log"

	"lewist543.com/mass-project-updater/internal/config"
	"lewist543.com/mass-project-updater/internal/gitlab"
	"lewist543.com/mass-project-updater/internal/packagejson"
	"lewist543.com/mass-project-updater/internal/runner"
)

func main() {
	cfg := config.Load()
	gitRunner := git.New()
	npmRunner := npm.New()

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

	for _, p := range projects {
		log.Println("Processing:", p.Name)
		if err := runner.RunProject(cfg, p, updates, gitRunner, npmRunner); err != nil {
			log.Println("FAILED:", p.Name, err)
		}
	}
}
