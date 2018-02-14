package linters

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/errors"
	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

func setup() TaskLinter {
	return TaskLinter{
		Fs: afero.Afero{Fs: afero.NewMemMapFs()},
	}
}
func TestAtLeastOneTaskExists(t *testing.T) {
	man := model.Manifest{}
	taskLinter := setup()

	errs := taskLinter.Lint(man)
	assert.Len(t, errs, 1)
	assert.IsType(t, errors.MissingField{}, errs[0])
}

func TestRunTaskWithoutScriptAndImage(t *testing.T) {
	man := model.Manifest{}
	taskLinter := setup()

	man.Tasks = []model.Task{
		model.Run{},
	}

	errs := taskLinter.Lint(man)
	assert.Len(t, errs, 3)

	assert.IsType(t, errors.MissingField{}, errs[0])
	assert.IsType(t, errors.MissingField{}, errs[1])
}

func TestRunTaskWithScriptAndImage(t *testing.T) {
	taskLinter := setup()
	man := model.Manifest{}
	man.Tasks = []model.Task{
		model.Run{
			Script: "./build.sh",
			Image:  "alpine",
		},
	}

	errs := taskLinter.Lint(man)
	assert.Len(t, errs, 1)
	assert.IsType(t, errors.FileError{}, errs[0])
}

func TestRunTaskScriptFileExists(t *testing.T) {
	taskLinter := setup()
	taskLinter.Fs.WriteFile("build.sh", []byte("foo"), 0777)

	man := model.Manifest{}
	man.Tasks = []model.Task{
		model.Run{
			Script: "./build.sh",
			Image:  "alpine",
		},
	}

	errs := taskLinter.Lint(man)
	assert.Len(t, errs, 0)
}