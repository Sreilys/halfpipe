package linters

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/linters/tasks"
	"github.com/springernature/halfpipe/manifest"
	"sort"
	"strings"
)

type taskLinter struct {
	Fs                              afero.Afero
	lintRunTask                     func(task manifest.Run, fs afero.Afero, os string) (errs []error, warnings []error)
	lintDeployCFTask                func(task manifest.DeployCF, fs afero.Afero) (errs []error, warnings []error)
	LintPrePromoteTask              func(task manifest.Task) (errs []error, warnings []error)
	lintDockerPushTask              func(task manifest.DockerPush, fs afero.Afero) (errs []error, warnings []error)
	lintDockerComposeTask           func(task manifest.DockerCompose, fs afero.Afero) (errs []error, warnings []error)
	lintConsumerIntegrationTestTask func(task manifest.ConsumerIntegrationTest, providerHostRequired bool) (errs []error, warnings []error)
	lintDeployMLZipTask             func(task manifest.DeployMLZip) (errs []error, warnings []error)
	lintDeployMLModulesTask         func(task manifest.DeployMLModules) (errs []error, warnings []error)
	lintArtifacts                   func(currentTask manifest.Task, previousTasks []manifest.Task) (errs []error, warnings []error)
	os                              string
}

func NewTasksLinter(fs afero.Afero, os string) taskLinter {
	return taskLinter{
		Fs:                              fs,
		lintRunTask:                     tasks.LintRunTask,
		lintDeployCFTask:                tasks.LintDeployCFTask,
		LintPrePromoteTask:              tasks.LintPrePromoteTask,
		lintDockerPushTask:              tasks.LintDockerPushTask,
		lintDockerComposeTask:           tasks.LintDockerComposeTask,
		lintConsumerIntegrationTestTask: tasks.LintConsumerIntegrationTestTask,
		lintDeployMLZipTask:             tasks.LintDeployMLZipTask,
		lintDeployMLModulesTask:         tasks.LintDeployMLModulesTask,
		lintArtifacts:                   tasks.LintArtifacts,
		os:                              os,
	}
}

func (linter taskLinter) Lint(man manifest.Manifest) (result result.LintResult) {
	result.Linter = "Tasks"
	result.DocsURL = "https://docs.halfpipe.io/manifest/#tasks"

	if len(man.Tasks) == 0 {
		result.AddError(errors.NewMissingField("tasks"))
		return
	}

	errs, warnings := linter.lintTasks("", man.Tasks, []manifest.Task{}, true)
	sortErrors(errs)
	sortErrors(warnings)

	result.AddError(errs...)
	result.AddWarning(warnings...)

	return
}

func (linter taskLinter) lintTasks(listName string, ts []manifest.Task, previousTasks []manifest.Task, lintArtifact bool) (rE []error, rW []error) {
	for index, t := range ts {
		previousTasks = append(previousTasks, ts[:index]...)

		var taskID string
		if listName == "" {
			taskID = fmt.Sprintf("tasks[%v]", index)
		} else {
			taskID = fmt.Sprintf("%s[%v]", listName, index)
		}

		prefixErrors := prefixErrorsWithIndex(taskID)

		var errs []error
		var warnings []error
		switch task := t.(type) {
		case manifest.Run:
			errs, warnings = linter.lintRunTask(task, linter.Fs, linter.os)
		case manifest.DeployCF:
			errs, warnings = linter.lintDeployCFTask(task, linter.Fs)

			if len(errs) == 0 && len(task.PrePromote) > 0 {
				for pI, preTask := range task.PrePromote {
					prePromotePrefixer := prefixErrorsWithIndex(fmt.Sprintf("%s.pre_promote[%v]", taskID, pI))
					e, w := linter.LintPrePromoteTask(preTask)
					errs = append(errs, prePromotePrefixer(e)...)
					warnings = append(warnings, prePromotePrefixer(w)...)
				}

				subErrors, subWarnings := linter.lintTasks(fmt.Sprintf("%s.pre_promote", taskID), task.PrePromote, previousTasks, false)
				errs = append(errs, subErrors...)
				warnings = append(warnings, subWarnings...)
			}
		case manifest.DockerPush:
			errs, warnings = linter.lintDockerPushTask(task, linter.Fs)
		case manifest.DockerCompose:
			errs, warnings = linter.lintDockerComposeTask(task, linter.Fs)
		case manifest.ConsumerIntegrationTest:
			if listName == "tasks" {
				errs, warnings = linter.lintConsumerIntegrationTestTask(task, true)
			} else {
				errs, warnings = linter.lintConsumerIntegrationTestTask(task, false)
			}
		case manifest.DeployMLZip:
			errs, warnings = linter.lintDeployMLZipTask(task)
		case manifest.DeployMLModules:
			errs, warnings = linter.lintDeployMLModulesTask(task)
		default:
			errs = append(errs, errors.NewInvalidField("task", fmt.Sprintf("%s is not a known task", taskID)))
		}

		if t.ReadsFromArtifacts() && lintArtifact {
			artifactErr, _ := linter.lintArtifacts(t, previousTasks)
			errs = append(errs, artifactErr...)
		}

		rE = append(rE, prefixErrors(errs)...)
		rW = append(rW, prefixErrors(warnings)...)
	}

	return
}

func sortErrors(errs []error) {
	getPrefix := func(err error) string {
		return strings.Split(err.Error(), " ")[0]
	}

	sort.Slice(errs, func(i, j int) bool {
		return getPrefix(errs[i]) < getPrefix(errs[j])
	})
}

func prefixErrorsWithIndex(prefix string) func(errs []error) (rE []error) {
	// Since we are calling lintTasks recursively we end up in a situation where
	// error already contains the prefix.
	return func(errs []error) (rE []error) {
		for _, e := range errs {
			if strings.HasPrefix(e.Error(), prefix) {
				rE = append(rE, e)
			} else {
				rE = append(rE, fmt.Errorf("%s %s", prefix, e))
			}

		}
		return
	}
}
