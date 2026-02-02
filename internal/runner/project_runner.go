package runner

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"

	"lewist543.com/mass-project-updater/internal/config"
	"lewist543.com/mass-project-updater/internal/git"
	"lewist543.com/mass-project-updater/internal/gitlab"
	"lewist543.com/mass-project-updater/internal/model"
	"lewist543.com/mass-project-updater/internal/npm"
	"lewist543.com/mass-project-updater/internal/packagejson"
)

// RunProject executes all stages for a single project.
// Writes full stage failures to ERROR_LOGS/<project-name>_<stage>.log
func RunProject(
	cfg config.Config,
	project model.Project,
	updates packagejson.Updates,
	gitRunner git.Git,
	npmRunner npm.NPM,
	mrClient gitlab.MRClient,
	projectDir string,
	logger *zap.Logger,
) (string, error) {

	dir := projectDir
	if dir == "" {
		dir = filepath.Join(cfg.WorkDir, project.Name)
	}

	logger.Info("Starting project update",
		zap.String("project", project.Name),
		zap.String("dir", dir),
		zap.Bool("dryRun", cfg.DryRun),
	)

	if cfg.DryRun {
		logger.Info("DryRun mode, skipping operations", zap.String("project", project.Name))
		return "", nil
	}

	// -------------------------
	// Git operations
	// -------------------------

	// Clone (needs URL)
	if err := runStage("clone", project, dir,
		func(d string, args ...interface{}) error {
			url := args[0].(string)
			return gitRunner.Clone(url, d)
		},
		logger, cfg.ErrorLogDir, project.SSHURL); err != nil {
		return "", err
	}

	// Checkout develop
	if err := runStage("checkout", project, dir,
		func(d string, args ...interface{}) error {
			return gitRunner.CheckoutDevelop(d)
		},
		logger, cfg.ErrorLogDir); err != nil {
		return "", err
	}

	// Create branch (needs branch name)
	if err := runStage("branch", project, dir,
		func(d string, args ...interface{}) error {
			branch := args[0].(string)
			return gitRunner.CreateBranch(d, branch)
		},
		logger, cfg.ErrorLogDir, cfg.BranchName); err != nil {
		return "", err
	}

	// -------------------------
	// Package.json update
	// -------------------------
	if err := runStage("packagejson", project, dir,
		func(d string, args ...interface{}) error {
			return packagejson.Apply(filepath.Join(d, "package.json"), updates)
		},
		logger, cfg.ErrorLogDir); err != nil {
		return "", err
	}

	// -------------------------
	// NPM operations
	// -------------------------
	if err := runStage("npminstall", project, dir,
		func(d string, args ...interface{}) error {
			return npmRunner.Install(d)
		},
		logger, cfg.ErrorLogDir); err != nil {
		return "", err
	}

	if err := runStage("npmbuild", project, dir,
		func(d string, args ...interface{}) error {
			return npmRunner.Build(d)
		},
		logger, cfg.ErrorLogDir); err != nil {
		return "", err
	}

	// -------------------------
	// Commit & Push
	// -------------------------
	if err := runStage("commit", project, dir,
		func(d string, args ...interface{}) error {
			return gitRunner.CommitAndPush(d, "chore: update dependencies", cfg.BranchName)
		},
		logger, cfg.ErrorLogDir); err != nil {
		return "", err
	}

	// -------------------------
	// Create MR
	// -------------------------
	mrURL, err := mrClient.CreateMR(project.ID, cfg.BranchName, "develop", "chore: update dependencies")
	if err != nil {
		logger.Error("MR creation failed", zap.String("project", project.Name), zap.String("stage", "create MR"), zap.Error(err))
		return "", fmt.Errorf("stage=create MR: %w", err)
	}
	logger.Info("MR created successfully", zap.String("project", project.Name), zap.String("mrURL", mrURL))

	return mrURL, nil
}

// -------------------------
// Helper: runStage
// -------------------------
func runStage(stage string, project model.Project, dir string, fn func(string, ...interface{}) error, logger *zap.Logger, errorLogDir string, args ...interface{}) error {
	if err := fn(dir, args...); err != nil {
		logFile := filepath.Join(errorLogDir, fmt.Sprintf("%s_%s.log", project.Name, stage))

		if execErr, ok := err.(interface{ Output() []byte }); ok {
			_ = os.WriteFile(logFile, execErr.Output(), 0644)

			logger.Error(fmt.Sprintf("%s failed (output written)", stage),
				zap.String("project", project.Name),
				zap.String("stage", stage),
				zap.String("logFile", logFile),
				zap.Error(err),
			)
		} else {
			logger.Error(fmt.Sprintf("%s failed", stage),
				zap.String("project", project.Name),
				zap.String("stage", stage),
				zap.Error(err),
			)
		}

		return fmt.Errorf("stage=%s: %w", stage, err)
	}

	logger.Info(fmt.Sprintf("%s succeeded", stage), zap.String("project", project.Name))
	return nil
}
