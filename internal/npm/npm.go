package npm

type NPM interface {
	Install(dir string) error
	Build(dir string) error
}
