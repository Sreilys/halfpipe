package pipeline

import (
	"testing"

	"regexp"

	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/model"
	"github.com/springernature/halfpipe/project"
	"github.com/stretchr/testify/assert"
)

func TestToString(t *testing.T) {
	man := model.Manifest{}
	man.Repo.Uri = "repo.git"

	actual, err := ToString(testPipeline().Render(project.Project{}, man))
	expected := "uri: repo.git"

	assert.Nil(t, err)
	assert.Contains(t, actual, expected)
}

func TestToStringVersionComment(t *testing.T) {
	man := model.Manifest{}
	man.Repo.Uri = "repo.git"
	config.Version = "0.0.1-yolo"

	actual, err := ToString(testPipeline().Render(project.Project{}, man))

	assert.Nil(t, err)
	assert.Regexp(t, regexp.MustCompile(`^#.*0\.0\.1-yolo.*`), actual)
}

func TestGeneratesUniqueNamesForJobsAndResources(t *testing.T) {
	manifest := model.Manifest{
		Repo: model.Repo{Uri: "https://github.com/springernature/halfpipe.git"},
		Tasks: []model.Task{
			model.Run{Script: "asd.sh"},
			model.Run{Script: "asd.sh"},
			model.Run{Script: "asd.sh"},
			model.Run{Script: "fgh.sh"},
			model.DeployCF{Org: "ee", Space: "dev"},
			model.DeployCF{Org: "ee", Space: "dev"},
			model.DeployCF{Org: "ee", Space: "dev"},
			model.DockerPush{},
			model.DockerPush{},
			model.DockerPush{},
		},
	}
	config := testPipeline().Render(project.Project{}, manifest)

	expectedJobNames := []string{
		"run asd.sh",
		"run asd.sh (1)",
		"run asd.sh (2)",
		"run fgh.sh",
		"deploy-cf",
		"deploy-cf (1)",
		"deploy-cf (2)",
		"docker-push",
		"docker-push (1)",
		"docker-push (2)",
	}

	expectedResourceNames := []string{
		"halfpipe",
		"CF ee-dev",
		"CF ee-dev (1)",
		"CF ee-dev (2)",
		"Docker Registry",
		"Docker Registry (1)",
		"Docker Registry (2)",
	}

	assert.Len(t, config.Jobs, len(expectedJobNames))
	assert.Len(t, config.Resources, len(expectedResourceNames))

	for i, name := range expectedJobNames {
		assert.Equal(t, name, config.Jobs[i].Name)
	}

	for i, name := range expectedResourceNames {
		assert.Equal(t, name, config.Resources[i].Name)
	}

}