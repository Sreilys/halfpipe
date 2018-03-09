package errors

import "fmt"

type MissingFieldError struct {
	Name string
}

func NewMissingField(name string) MissingFieldError {
	return MissingFieldError{name}
}

func (e MissingFieldError) Error() string {
	return fmt.Sprintf("Missing field: %s", e.Name)
}

func (e MissingFieldError) DocID() string {
	return "missing-field-" + e.Name
}
