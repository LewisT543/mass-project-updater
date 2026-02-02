package runner

import (
	"errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	_ "go.uber.org/zap/zaptest/observer"

	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"lewist543.com/mass-project-updater/internal/config"
	"lewist543.com/mass-project-updater/internal/model"
	"lewist543.com/mass-project-updater/internal/packagejson"
)

// ---------------------------
// Fakes
// ---------------------------

// Concurrency-safe FakeGit
type FakeGit struct {
	mu    sync.Mutex
	Calls map[string][]string // key: project name, value: git calls
}

// helper to map path -> project name for testing
func (f *FakeGit) record(dir, action string) {
	project := filepath.Base(dir) // extract last folder name
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Calls == nil {
		f.Calls = map[string][]string{}
	}
	f.Calls[project] = append(f.Calls[project], action)
}

func (f *FakeGit) Clone(_url, dir string) error {
	f.record(dir, "clone")
	return nil
}

func (f *FakeGit) CheckoutDevelop(dir string) error {
	f.record(dir, "checkout")
	return nil
}

func (f *FakeGit) CreateBranch(dir, _ string) error {
	f.record(dir, "branch")
	return nil
}

func (f *FakeGit) CommitAndPush(dir, _, _ string) error {
	f.record(dir, "commit")
	return nil
}

// Fake NPM runner
type FakeNPM struct {
	InstallCalled bool
	BuildCalled   bool
}

func (n *FakeNPM) Install(_ string) error { n.InstallCalled = true; return nil }
func (n *FakeNPM) Build(_ string) error   { n.BuildCalled = true; return nil }

// Fake MR client
type FakeMRClient struct {
	Called map[int]bool
}

func (f *FakeMRClient) CreateMR(projectID int, _, _, _ string) (string, error) {
	if f.Called == nil {
		f.Called = map[int]bool{}
	}
	f.Called[projectID] = true
	return "https://fake.gitlab.com/mr/" + string(rune(projectID)), nil
}

// ---------------------------
// Tests
// ---------------------------

func TestRunAllProjects_Concurrent(t *testing.T) {
	fakeGit := &FakeGit{}
	fakeNPM := &FakeNPM{}
	fakeMR := &FakeMRClient{}
	logger := zap.NewNop()

	cfg := config.Config{
		WorkDir:    t.TempDir(),
		BranchName: "chore/deps",
		DryRun:     false,
	}

	projects := []model.Project{
		{ID: 1, Name: "ui-spa-test1", SSHURL: "git@example.com:test1.git"},
		{ID: 2, Name: "ui-spa-test2", SSHURL: "git@example.com:test2.git"},
		{ID: 3, Name: "ui-spa-test3", SSHURL: "git@example.com:test3.git"},
	}

	updates := packagejson.Updates{
		Dependencies: map[string]string{
			"react":  "^18.3.0",
			"lodash": "^4.17.21", // ignored
		},
		DevDependencies: map[string]string{
			"vite": "^4.1.0",
			"jest": "^29.0.0", // ignored
		},
	}

	// Run projects concurrently
	results := RunAllProjects(projects, 2, func(p model.Project) (string, error) {
		dir := filepath.Join(cfg.WorkDir, p.Name)
		os.MkdirAll(dir, 0755)
		os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{
			"name": "`+p.Name+`",
			"version": "1.0.0",
			"dependencies": {"react": "^18.2.0","axios":"^1.3.0"},
			"devDependencies": {"vite":"^4.0.0"}
		}`), 0644)
		return RunProject(cfg, p, updates, fakeGit, fakeNPM, fakeMR, dir, logger)
	})

	// Validate all results
	for i, r := range results {
		assert.NoError(t, r.Err, "Project %s should succeed", projects[i].Name)
		assert.NotEmpty(t, r.MRURL)
		assert.True(t, fakeMR.Called[projects[i].ID])
	}

	// Validate git calls per project (concurrency-safe)
	for _, p := range projects {
		calls := fakeGit.Calls[p.Name]
		assert.Equal(t, []string{"clone", "checkout", "branch", "commit"}, calls, "Project %s git calls mismatch", p.Name)
	}

	assert.True(t, fakeNPM.InstallCalled)
	assert.True(t, fakeNPM.BuildCalled)

	// Validate snapshots for one project
	dir := filepath.Join(cfg.WorkDir, projects[0].Name)
	afterB, _ := os.ReadFile(filepath.Join(dir, "package.json"))
	expectedJSON := `{
  "name": "ui-spa-test1",
  "version": "1.0.0",
  "dependencies": {
    "react": "^18.3.0",
    "axios": "^1.3.0"
  },
  "devDependencies": {
    "vite": "^4.1.0"
  }
}`
	assert.JSONEq(t, expectedJSON, string(afterB))
}

func TestRunAllProjects_DryRun(t *testing.T) {
	fakeGit := &FakeGit{}
	fakeNPM := &FakeNPM{}
	fakeMR := &FakeMRClient{}
	logger := zap.NewNop()

	cfg := config.Config{
		WorkDir:    t.TempDir(),
		BranchName: "chore/deps",
		DryRun:     true,
	}

	project := model.Project{ID: 1, Name: "ui-spa-test", SSHURL: "git@example.com:test.git"}
	dir := filepath.Join(cfg.WorkDir, project.Name)
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies":{"react":"^18.2.0"}}`), 0644)

	results := RunAllProjects([]model.Project{project}, 1, func(p model.Project) (string, error) {
		return RunProject(cfg, p, packagejson.Updates{
			Dependencies: map[string]string{"react": "^18.3.0"},
		}, fakeGit, fakeNPM, fakeMR, dir, logger)
	})

	// Validate result
	assert.Len(t, results, 1)
	assert.Empty(t, results[0].MRURL)
	assert.NoError(t, results[0].Err)

	// Nothing should have run
	assert.Empty(t, fakeGit.Calls)
	assert.False(t, fakeNPM.InstallCalled)
	assert.False(t, fakeNPM.BuildCalled)
	assert.Empty(t, fakeMR.Called, "MR should not be called in dry-run mode")

	// Package.json unchanged
	afterB, _ := os.ReadFile(filepath.Join(dir, "package.json"))
	assert.JSONEq(t, `{"dependencies":{"react":"^18.2.0"}}`, string(afterB))
}

type fakeExecError struct {
	output []byte
}

func (e *fakeExecError) Error() string {
	return "command failed"
}

func (e *fakeExecError) Output() []byte {
	return e.output
}

// ---------------------------
// runStage tests
// ---------------------------

func TestRunStage_RegularError_LoggedNotWritten(t *testing.T) {
	errorDir := t.TempDir()

	core, logs := observer.New(zap.ErrorLevel)
	logger := zap.New(core)

	project := model.Project{Name: "ui-spa-test"}

	err := runStage(
		"npmbuild",
		project,
		"/tmp/project",
		func(_ string, _ ...interface{}) error {
			return errors.New("boom")
		},
		logger,
		errorDir,
	)

	assert.Error(t, err)

	// Assert log written
	assert.Len(t, logs.All(), 1)
	assert.Equal(t, "npmbuild failed", logs.All()[0].Message)

	// Assert no files written
	entries, _ := os.ReadDir(errorDir)
	assert.Len(t, entries, 0)
}

func TestRunStage_ExecError_WrittenToFile(t *testing.T) {
	errorDir := t.TempDir()

	core, logs := observer.New(zap.ErrorLevel)
	logger := zap.New(core)

	project := model.Project{Name: "ui-spa-test"}

	execErr := &fakeExecError{
		output: []byte("npm ERR! build failed"),
	}

	err := runStage(
		"npmbuild",
		project,
		"/tmp/project",
		func(_ string, _ ...interface{}) error {
			return execErr
		},
		logger,
		errorDir,
	)

	assert.Error(t, err)

	// Assert log written
	assert.Len(t, logs.All(), 1)
	assert.Equal(t, "npmbuild failed", logs.All()[0].Message)

	// Assert file written
	files, _ := os.ReadDir(errorDir)
	assert.Len(t, files, 1)

	logFile := filepath.Join(errorDir, files[0].Name())
	content, _ := os.ReadFile(logFile)

	assert.Contains(t, string(content), "npm ERR! build failed")
}
