package packagejson

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func load(path string) string {
	b, _ := os.ReadFile(path)
	return string(b)
}

func TestApply_Snapshot(t *testing.T) {
	dir := t.TempDir()
	pkgPath := filepath.Join(dir, "package.json")
	os.WriteFile(pkgPath, []byte(load("testdata/before.package.json")), 0644)

	updates := Updates{
		Dependencies: map[string]string{
			"react": "^18.3.0",
		},
	}

	Apply(pkgPath, updates)

	actual := load(pkgPath)
	expected := load("testdata/after.package.json")
	assert.JSONEq(t, expected, actual)
}
