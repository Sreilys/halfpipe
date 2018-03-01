package linters

import (
	"testing"

	"fmt"

	"github.com/springernature/halfpipe/errors"
	"github.com/stretchr/testify/assert"
)

func assertMissingField(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(errors.MissingFieldError)
	if !ok {
		assert.Fail(t, "error is not a MissingField", err)
	} else {
		assert.Equal(t, name, mf.Name)
	}
}

func assertInvalidField(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(errors.InvalidFieldError)
	if !ok {
		assert.Fail(t, "error is not an InvalidField", err)
	} else {
		assert.Equal(t, name, mf.Name)
	}
}

func assertInvalidFieldInErrors(t *testing.T, name string, errs []error) {
	t.Helper()

	for _, err := range errs {
		mf, ok := err.(errors.InvalidFieldError)
		if ok {
			if mf.Name == name {
				return
			}
		}
	}
	assert.Fail(t, fmt.Sprintf("Could not find invalid field error for '%s' in %s", name, errs))
}

func assertInvalidFieldShouldNotBeInErrors(t *testing.T, name string, errs []error) {
	t.Helper()

	for _, err := range errs {
		mf, ok := err.(errors.InvalidFieldError)
		if ok {
			if mf.Name == name {
				assert.Fail(t, fmt.Sprintf("Invalid field error for '%s' should not be in %s!", name, errs))
			}
		}
	}
}

func assertFileError(t *testing.T, path string, err error) {
	t.Helper()

	mf, ok := err.(errors.FileError)
	if !ok {
		assert.Fail(t, "error is not a FileError", err)
	} else {
		assert.Equal(t, path, mf.Path)
	}
}