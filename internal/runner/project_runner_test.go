package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"lewist543.com/mass-project-updater/internal/config"
	"lewist543.com/mass-project-updater/internal/model"
	"lewist543.com/mass-project-updater/internal/packagejson"
)

func TestRunProject_HappyPath(t *testing.T) {
	fakeGit := &FakeGit{}
	fakeNpm := &FakeNPM{}

	cfg := config.Config{
		WorkDir:    t.TempDir(),
		BranchName: "chore/deps",
	}

	project := model.Project{
		Name:   "ui-spa-test",
		SSHURL: "fakeGit@example.com:test.fakeGit",
	}

	updates := packagejson.Updates{}

	err := RunProject(cfg, project, updates, fakeGit, fakeNpm)

	assert.NoError(t, err)
	assert.True(t, fakeNpm.InstallCalled)
	assert.True(t, fakeNpm.BuildCalled)

	assert.Equal(t, []string{
		"clone",
		"checkout",
		"branch",
		"commit",
	}, fakeGit.Calls)
}
