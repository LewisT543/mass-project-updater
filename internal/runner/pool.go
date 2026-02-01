package runner

import (
	"sync"

	"lewist543.com/mass-project-updater/internal/model"
)

type ProjectResult struct {
	Project model.Project
	MRURL   string
	Err     error
}

// RunAllProjects runs multiple projects concurrently using a worker pool.
// `workers` sets the maximum number of parallel executions.
// `runFunc` is the function to execute per project.
func RunAllProjects(
	projects []model.Project,
	workers int,
	runFunc func(model.Project) (string, error),
) []ProjectResult {
	var wg sync.WaitGroup
	sem := make(chan struct{}, workers)
	results := make([]ProjectResult, len(projects))

	for i, project := range projects {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, p model.Project) {
			defer wg.Done()
			defer func() { <-sem }()

			mrUrl, err := runFunc(p)

			results[idx] = ProjectResult{
				Project: p,
				MRURL:   mrUrl,
				Err:     err,
			}
		}(i, project)
	}

	wg.Wait()
	return results
}
