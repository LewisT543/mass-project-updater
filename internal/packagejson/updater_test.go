package packagejson

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApply_UpdatesOnlyExistingDeps(t *testing.T) {
	dir := t.TempDir()
	pkgPath := dir + "/package.json"

	os.WriteFile(pkgPath, []byte(`{
      "dependencies": {
        "react": "^18.2.0"
      },
      "devDependencies": {
        "vite": "^4.0.0"
      }
    }`), 0644)

	updates := Updates{
		Dependencies: map[string]string{
			"react":  "^18.3.0",
			"lodash": "^4.17.0",
		},
		DevDependencies: map[string]string{
			"vite": "^5.0.0",
		},
	}

	err := Apply(pkgPath, updates)
	assert.NoError(t, err)

	out, _ := os.ReadFile(pkgPath)

	assert.Contains(t, string(out), `"react": "^18.3.0"`)
	assert.Contains(t, string(out), `"vite": "^5.0.0"`)
	assert.NotContains(t, string(out), `"lodash"`)
}
