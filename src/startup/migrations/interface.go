package migrations

type Migration interface {
	Apply() error
}
