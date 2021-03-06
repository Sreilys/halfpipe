package pipeline

import (
	"testing"

	"regexp"

	con "github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestToString(t *testing.T) {
	man := manifest.Manifest{}
	man.Repo.URI = "repo.git"

	actual, err := ToString(testPipeline().Render(man))
	expected := "uri: repo.git"

	assert.Nil(t, err)
	assert.Contains(t, actual, expected)
}

func TestToStringVersionComment(t *testing.T) {
	man := manifest.Manifest{}
	man.Repo.URI = "repo.git"
	con.Version = "0.0.1-yolo"

	actual, err := ToString(testPipeline().Render(man))

	assert.Nil(t, err)
	assert.Regexp(t, regexp.MustCompile(`^#.*0\.0\.1-yolo.*`), actual)
}

func TestGeneratesUniqueNamesForJobsAndResources(t *testing.T) {
	man := manifest.Manifest{
		Repo: manifest.Repo{URI: "https://github.com/springernature/halfpipe.git"},
		Tasks: []manifest.Task{
			manifest.Run{Script: "asd.sh"},
			manifest.Run{Script: "asd.sh"},
			manifest.Run{Name: "test", Script: "asd.sh"},
			manifest.Run{Name: "test", Script: "fgh.sh"},
			manifest.DeployCF{API: "api.foo.bar", Org: "ee", Space: "dev"},
			manifest.DeployCF{API: "https://api.foo.bar", Org: "ee", Space: "dev"},
			manifest.DeployCF{API: "((cloudfoundry.api-dev))", Org: "((cloudfoundry.org-dev))", Space: "((cloudfoundry.space-dev))"},
			manifest.DockerPush{},
			manifest.DockerPush{},
			manifest.DockerPush{},
			manifest.DeployCF{Name: "deploy to dev"},
			manifest.DeployCF{Name: "deploy to dev"},
			manifest.DockerPush{Name: "push to docker hub"},
			manifest.DockerPush{Name: "push to docker hub"},
		},
	}
	config := testPipeline().Render(man)

	expectedJobNames := []string{
		"run asd.sh",
		"run asd.sh (1)",
		"test",
		"test (1)",
		"deploy-cf",
		"deploy-cf (1)",
		"deploy-cf (2)",
		"docker-push",
		"docker-push (1)",
		"docker-push (2)",
		"deploy to dev",
		"deploy to dev (1)",
		"push to docker hub",
		"push to docker hub (1)",
	}

	expectedResourceNames := []string{
		gitDir,
		"CF api.foo.bar ee dev",
		"CF api.foo.bar ee dev (1)",
		"CF dev org-dev space-dev",
		"Docker Registry",
		"Docker Registry (1)",
		"Docker Registry (2)",
		"CF   ",
		"CF    (1)",
		"Docker Registry (3)",
		"Docker Registry (4)",
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
