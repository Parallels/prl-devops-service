package interfaces

type DatabaseContext interface {
	Connect() error
}
