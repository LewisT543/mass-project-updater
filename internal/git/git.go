package git

type Git interface {
	Clone(url, dir string) error
	CheckoutDevelop(dir string) error
	CreateBranch(dir, name string) error
	CommitAndPush(dir, msg, branch string) error
}
