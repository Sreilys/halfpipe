package pipeline

import (
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRenderDockerPushTask(t *testing.T) {
	man := manifest.Manifest{}
	man.Repo.URI = "git@github.com:/springernature/foo.git"

	username := "halfpipe"
	password := "secret"
	repo := "halfpipe/halfpipe-cli"
	man.Tasks = []manifest.Task{
		manifest.DockerPush{
			Username: username,
			Password: password,
			Image:    repo,
			Vars: manifest.Vars{
				"A": "a",
				"B": "b",
			},
		},
	}

	expectedResource := atc.ResourceConfig{
		Name: "Docker Registry",
		Type: "docker-image",
		Source: atc.Source{
			"username":   username,
			"password":   password,
			"repository": repo,
		},
	}

	expectedJobConfig := atc.JobConfig{
		Name:   "docker-push",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Aggregate: &atc.PlanSequence{atc.PlanConfig{Get: gitDir, Trigger: true}}},
			atc.PlanConfig{
				Attempts: 1,
				Put:      "Docker Registry",
				Params: atc.Params{
					"build": gitDir,
					"build_args": map[string]interface{}{
						"A": "a",
						"B": "b",
					},
					"tag_as_latest": true,
				},
			},
		},
	}

	// First resource will always be the git resource.
	assert.Equal(t, expectedResource, testPipeline().Render(man).Resources[1])
	assert.Equal(t, expectedJobConfig, testPipeline().Render(man).Jobs[0])
}

func TestRenderDockerPushTaskNotInRoot(t *testing.T) {
	man := manifest.Manifest{}
	man.Repo.URI = "git@github.com:/springernature/foo.git"
	basePath := "subapp/sub2"
	man.Repo.BasePath = basePath

	username := "halfpipe"
	password := "secret"
	repo := "halfpipe/halfpipe-cli"
	man.Tasks = []manifest.Task{
		manifest.DockerPush{
			Username: username,
			Password: password,
			Image:    repo,
		},
	}

	expectedResource := atc.ResourceConfig{
		Name: "Docker Registry",
		Type: "docker-image",
		Source: atc.Source{
			"username":   username,
			"password":   password,
			"repository": repo,
		},
	}

	expectedJobConfig := atc.JobConfig{
		Name:   "docker-push",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Aggregate: &atc.PlanSequence{atc.PlanConfig{Get: gitDir, Trigger: true}}},
			atc.PlanConfig{
				Attempts: 1,
				Put:      "Docker Registry",
				Params: atc.Params{
					"build":         gitDir + "/" + basePath,
					"tag_as_latest": true,
				}},
		},
	}

	// First resource will always be the git resource.
	assert.Equal(t, expectedResource, testPipeline().Render(man).Resources[1])
	assert.Equal(t, expectedJobConfig, testPipeline().Render(man).Jobs[0])
}

func TestRenderDockerPushWithVersioning(t *testing.T) {
	basePath := "subapp/sub2"
	man := manifest.Manifest{
		Repo: manifest.Repo{
			URI:      "git@github.com:/springernature/foo.git",
			BasePath: basePath,
		},
		FeatureToggles: manifest.FeatureToggles{
			manifest.FeatureVersioned,
		},
	}

	username := "halfpipe"
	password := "secret"
	repo := "halfpipe/halfpipe-cli"
	man.Tasks = []manifest.Task{
		manifest.DockerPush{
			Username: username,
			Password: password,
			Image:    repo,
		},
	}

	expectedResource := atc.ResourceConfig{
		Name: "Docker Registry",
		Type: "docker-image",
		Source: atc.Source{
			"username":   username,
			"password":   password,
			"repository": repo,
		},
	}

	expectedJobConfig := atc.JobConfig{
		Name:   "docker-push",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{
				Aggregate: &atc.PlanSequence{
					atc.PlanConfig{Get: gitDir, Passed: []string{"update version"}},
					atc.PlanConfig{Get: versionName, Passed: []string{"update version"}, Trigger: true},
				},
			},
			atc.PlanConfig{
				Attempts: 1,
				Put:      "Docker Registry",
				Params: atc.Params{
					"tag_file":      "version/number",
					"build":         gitDir + "/" + basePath,
					"tag_as_latest": true,
				}},
		},
	}

	// First resource will always be the git resource.
	assert.Equal(t, expectedResource, testPipeline().Render(man).Resources[2])
	assert.Equal(t, expectedJobConfig, testPipeline().Render(man).Jobs[1])
}

func TestRenderDockerPushWithVersioningAndRestoreArtifact(t *testing.T) {
	basePath := "subapp/sub2"
	man := manifest.Manifest{
		Repo: manifest.Repo{
			URI:      "git@github.com:/springernature/foo.git",
			BasePath: basePath,
		},
		FeatureToggles: manifest.FeatureToggles{
			manifest.FeatureVersioned,
		},
	}

	username := "halfpipe"
	password := "secret"
	repo := "halfpipe/halfpipe-cli"
	man.Tasks = []manifest.Task{
		manifest.DockerPush{
			Username:         username,
			Password:         password,
			Image:            repo,
			RestoreArtifacts: true,
		},
	}

	expectedResource := atc.ResourceConfig{
		Name: "Docker Registry",
		Type: "docker-image",
		Source: atc.Source{
			"username":   username,
			"password":   password,
			"repository": repo,
		},
	}

	jobName := "docker-push"
	expectedJobConfig := atc.JobConfig{
		Name:   jobName,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{
				Aggregate: &atc.PlanSequence{
					atc.PlanConfig{Get: gitDir, Passed: []string{"update version"}},
					atc.PlanConfig{Get: versionName, Passed: []string{"update version"}, Trigger: true},
				},
			},
			restoreArtifactTask(man),
			atc.PlanConfig{
				Task: "Copying git repo and artifacts to a temporary build dir",
				TaskConfig: &atc.TaskConfig{
					Platform: "linux",
					ImageResource: &atc.ImageResource{
						Type: "docker-image",
						Source: atc.Source{
							"repository": "alpine",
						},
					},
					Run: atc.TaskRunConfig{
						Path: "/bin/sh",
						Args: []string{"-c", strings.Join([]string{
							fmt.Sprintf("cp -r %s/. %s", gitDir, dockerBuildTmpDir),
							fmt.Sprintf("cp -r %s/. %s", artifactsInDir, path.Join(dockerBuildTmpDir, man.Repo.BasePath)),
						}, "\n")},
					},
					Inputs: []atc.TaskInputConfig{
						{Name: gitDir},
						{Name: artifactsName},
					},
					Outputs: []atc.TaskOutputConfig{
						{Name: dockerBuildTmpDir},
					},
				},
			},
			atc.PlanConfig{
				Attempts: 1,
				Put:      "Docker Registry",
				Params: atc.Params{
					"tag_file":      "version/number",
					"build":         dockerBuildTmpDir + "/" + basePath,
					"tag_as_latest": true,
				}},
		},
	}

	// First resource will always be the git resource.
	dockerResource, found := testPipeline().Render(man).Resources.Lookup(dockerPushResourceName)
	assert.True(t, found)
	assert.Equal(t, expectedResource, dockerResource)

	config, foundJob := testPipeline().Render(man).Jobs.Lookup(jobName)
	assert.True(t, foundJob)
	assert.Equal(t, expectedJobConfig, config)
}
