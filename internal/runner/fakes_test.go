package runner

type FakeGit struct {
	Calls []string
	Fail  bool
}

func (f *FakeGit) Clone(_, _ string) error {
	f.Calls = append(f.Calls, "clone")
	return nil
}
func (f *FakeGit) CheckoutDevelop(_ string) error {
	f.Calls = append(f.Calls, "checkout")
	return nil
}
func (f *FakeGit) CreateBranch(_, _ string) error {
	f.Calls = append(f.Calls, "branch")
	return nil
}
func (f *FakeGit) CommitAndPush(_, _, _ string) error {
	f.Calls = append(f.Calls, "commit")
	return nil
}

type FakeNPM struct {
	InstallCalled bool
	BuildCalled   bool
}

func (n *FakeNPM) Install(_ string) error {
	n.InstallCalled = true
	return nil
}
func (n *FakeNPM) Build(_ string) error {
	n.BuildCalled = true
	return nil
}
