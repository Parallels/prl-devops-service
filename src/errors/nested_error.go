package errors

type NestedError struct {
	Message     string
	Path        string
	Code        int
	Description string
}
