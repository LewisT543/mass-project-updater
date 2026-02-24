package packagejson

import (
	"encoding/json"
	"os"
)

type Updates struct {
	Dependencies    map[string]string      `json:"dependencies"`
	DevDependencies map[string]string      `json:"devDependencies"`
	Overrides       map[string]interface{} `json:"overrides"`
}

func LoadUpdates(path string) (Updates, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Updates{}, err
	}
	var u Updates
	err = json.Unmarshal(b, &u)
	return u, err
}

func Apply(pkgPath string, updates Updates) error {
	b, err := os.ReadFile(pkgPath)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	// ------------------------
	// Update existing dependencies
	// ------------------------
	if deps, ok := data["dependencies"].(map[string]interface{}); ok {
		for k, v := range updates.Dependencies {
			if _, exists := deps[k]; exists {
				deps[k] = v
			}
		}
		data["dependencies"] = deps
	}

	if devDeps, ok := data["devDependencies"].(map[string]interface{}); ok {
		for k, v := range updates.DevDependencies {
			if _, exists := devDeps[k]; exists {
				devDeps[k] = v
			}
		}
		data["devDependencies"] = devDeps
	}

	// ------------------------
	// Apply overrides
	// ------------------------
	if len(updates.Overrides) > 0 {
		current, _ := data["overrides"].(map[string]interface{})
		if current == nil {
			current = map[string]interface{}{}
		}

		for k, v := range updates.Overrides {
			current[k] = v
		}

		data["overrides"] = current
	}

	// Write back
	newB, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(pkgPath, newB, 0644)
}
