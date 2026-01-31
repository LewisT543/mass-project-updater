package packagejson

import (
	"encoding/json"
	"os"
)

type Updates struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

type PackageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
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

func Apply(path string, updates Updates) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var pkg PackageJSON
	json.Unmarshal(b, &pkg)

	for d, v := range updates.Dependencies {
		if _, ok := pkg.Dependencies[d]; ok {
			pkg.Dependencies[d] = v
		}
	}

	for d, v := range updates.DevDependencies {
		if _, ok := pkg.DevDependencies[d]; ok {
			pkg.DevDependencies[d] = v
		}
	}

	out, _ := json.MarshalIndent(pkg, "", "  ")
	return os.WriteFile(path, out, 0644)
}
