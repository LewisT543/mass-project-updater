package gitlab

import "fmt"

type FakeMRClient struct {
	Called bool
}

func (f *FakeMRClient) CreateMR(projectID int, sourceBranch, targetBranch, title string) (string, error) {
	f.Called = true
	return fmt.Sprintf("https://fake.gitlab.com/mr/%d", projectID), nil
}
