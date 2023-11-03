package interfaces

type Service interface {
	Name() string
	FindPath() string
	Version() string
	Install(asUser, version string, flags map[string]string) error
	Uninstall(asUser string, uninstallDependencies bool) error
	Installed() bool
	Dependencies() []Service
	SetDependencies(dependencies []Service)
}
