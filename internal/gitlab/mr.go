package gitlab

type MRClient interface {
	CreateMR(projectID int, sourceBranch, targetBranch, title string) (string, error)
}
