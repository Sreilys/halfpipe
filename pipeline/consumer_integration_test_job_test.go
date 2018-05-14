package pipeline

import (
	"strings"
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRenderConsumerIntegrationTestTask(t *testing.T) {
	p := testPipeline()

	man := manifest.Manifest{
		Pipeline: "p-name",
		Repo: manifest.Repo{
			URI:      "git@git:user/repo",
			BasePath: "base.path",
		},
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Name:  "cf-deploy",
				API:   "cf-api",
				Space: "cf-space",
				Org:   "cf-org",
				PrePromote: []manifest.Task{
					manifest.ConsumerIntegrationTest{
						Name:     "c-name",
						Consumer: "c-consumer/c-path",
						Host:     "c-host",
						Script:   "c-script",
					},
				},
			},
		},
	}

	expectedVars := map[string]string{
		"CONSUMER_GIT_URI":  "git@github.com:springernature/c-consumer",
		"CONSUMER_PATH":     "c-path",
		"CONSUMER_SCRIPT":   "c-script",
		"CONSUMER_GIT_KEY":  "((github.private_key))",
		"CONSUMER_HOST":     "c-host",
		"PROVIDER_NAME":     "p-name",
		"PROVIDER_HOST_KEY": "P_NAME_DEPLOYED_HOST",
		"PROVIDER_HOST":     buildTestRoute("test-name", "cf-space", ""),
	}

	expectedJob := atc.JobConfig{
		Name:   "c-name",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{
				Get:     gitDir,
				Trigger: true,
				Passed:  []string{"cf-deploy"},
			},
			atc.PlanConfig{
				Task:       "run",
				Privileged: true,
				TaskConfig: &atc.TaskConfig{
					Platform: "linux",
					Params:   expectedVars,
					ImageResource: &atc.ImageResource{
						Type: "docker-image",
						Source: atc.Source{
							"repository": strings.Split(config.ConsumerIntegrationTestImage, ":")[0],
							"tag":        strings.Split(config.ConsumerIntegrationTestImage, ":")[1],
							"username":   "_json_key",
							"password":   "((gcr.private_key))",
						},
					},
					Run: atc.TaskRunConfig{
						Path: "/bin/sh",
						Dir:  gitDir + "/base.path",
						Args: runScriptArgs(consumerIntegrationTestScript, "", nil, "../.git/ref"),
					},
					Inputs: []atc.TaskInputConfig{
						{Name: gitDir},
					},
				}},
		},
	}

	jobs := p.Render(man).Jobs
	if assert.Len(t, jobs, 3) {
		assert.Equal(t, expectedJob, jobs[1])
	}
}