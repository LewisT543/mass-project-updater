package runner

import (
	"path/filepath"

	"lewist543.com/mass-project-updater/internal/config"
	"lewist543.com/mass-project-updater/internal/git"
	"lewist543.com/mass-project-updater/internal/model"
	"lewist543.com/mass-project-updater/internal/npm"
	"lewist543.com/mass-project-updater/internal/packagejson"
)

func RunProject(
	cfg config.Config,
	project model.Project,
	updates packagejson.Updates,
	git git.Git,
	npm npm.NPM,
) error {

	dir := filepath.Join(cfg.WorkDir, project.Name)

	if err := git.Clone(project.SSHURL, dir); err != nil {
		return err
	}

	if err := git.CheckoutDevelop(dir); err != nil {
		return err
	}

	if err := git.CreateBranch(dir, cfg.BranchName); err != nil {
		return err
	}

	if err := packagejson.Apply(
		filepath.Join(dir, "package.json"),
		updates,
	); err != nil {
		return err
	}

	if err := npm.Install(dir); err != nil {
		return err
	}

	if err := npm.Build(dir); err != nil {
		return err
	}

	return git.CommitAndPush(
		dir,
		"chore: update dependencies",
		cfg.BranchName,
	)
}
