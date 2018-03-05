package pipeline

import (
	"fmt"
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRendersHttpGitResource(t *testing.T) {
	name := "yolo"
	gitUri := fmt.Sprintf("git@github.com:springernature/%s.git", name)

	man := manifest.Manifest{}
	man.Repo.Uri = gitUri

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: name,
				Type: "git",
				Source: atc.Source{
					"uri": gitUri,
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(man))
}

func TestRendersSshGitResource(t *testing.T) {
	name := "asdf"
	gitUri := fmt.Sprintf("git@github.com:springernature/%s.git/", name)
	privateKey := "blurgh"

	man := manifest.Manifest{}
	man.Repo.Uri = gitUri
	man.Repo.PrivateKey = privateKey

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: name,
				Type: "git",
				Source: atc.Source{
					"uri":         gitUri,
					"private_key": privateKey,
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(man))
}

func TestRendersGitResourceWithWatchesAndIgnores(t *testing.T) {
	name := "asdf"
	gitUri := fmt.Sprintf("git@github.com:springernature/%s.git/", name)
	privateKey := "blurgh"

	man := manifest.Manifest{}
	man.Repo.Uri = gitUri
	man.Repo.PrivateKey = privateKey

	watches := []string{"watch1", "watch2"}
	ignores := []string{"ignore1", "ignore2"}
	man.Repo.WatchedPaths = watches
	man.Repo.IgnoredPaths = ignores

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: name,
				Type: "git",
				Source: atc.Source{
					"uri":          gitUri,
					"private_key":  privateKey,
					"paths":        watches,
					"ignore_paths": ignores,
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(man))
}

func TestRendersHttpGitResourceWithGitCrypt(t *testing.T) {
	name := "yolo"
	gitUri := fmt.Sprintf("git@github.com:springernature/%s.git", name)
	gitCrypt := "AABBFF66"

	man := manifest.Manifest{}
	man.Repo.Uri = gitUri
	man.Repo.GitCryptKey = gitCrypt

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: name,
				Type: "git",
				Source: atc.Source{
					"uri":           gitUri,
					"git_crypt_key": gitCrypt,
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(man))
}