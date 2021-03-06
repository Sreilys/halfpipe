package manifest_test

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

var secretValidator = manifest.NewSecretValidator()

func TestTopLevelManifest(t *testing.T) {

	bad := manifest.Manifest{
		Team:            "something ((secret.stuff))",
		Pipeline:        "((kehe.kehu))",
		SlackChannel:    "((asdf.dsa))",
		TriggerInterval: "((my.secret))",
	}

	errors := secretValidator.Validate(bad)
	assert.Len(t, errors, 4)
	assert.Contains(t, errors, manifest.UnsupportedSecretError("team"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("pipeline"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("slack_channel"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("trigger_interval"))

	good := manifest.Manifest{
		Team:            "a",
		Pipeline:        "b",
		SlackChannel:    "c",
		TriggerInterval: "d",
	}

	assert.Empty(t, secretValidator.Validate(good))
}

func TestRepo(t *testing.T) {
	bad := manifest.Manifest{
		Repo: manifest.Repo{
			URI:        "((not.allowed))",
			BasePath:   "((not.allowed))",
			PrivateKey: "((allowed.yo))",
			WatchedPaths: []string{
				"ok",
				"((not.allowed))",
			},
			IgnoredPaths: []string{
				"ok",
				"okAgain",
				"((not.allowed))",
			},
			GitCryptKey: "((allowed.yo))",
			Branch:      "((not.allowed))",
		},
	}

	errors := secretValidator.Validate(bad)
	assert.Len(t, errors, 5)
	assert.Contains(t, errors, manifest.UnsupportedSecretError("repo.uri"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("repo.basePath"))
	assert.NotContains(t, errors, manifest.UnsupportedSecretError("repo.private_key"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("repo.watched_paths[1]"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("repo.ignored_paths[2]"))
	assert.NotContains(t, errors, manifest.UnsupportedSecretError("repo.git_crypt_key"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("repo.branch"))

	good := manifest.Manifest{
		Repo: manifest.Repo{
			URI:          "Kehe",
			BasePath:     "Kehu,",
			PrivateKey:   "((super.allowed))",
			WatchedPaths: []string{"a", "b"},
			IgnoredPaths: []string{"d", "e"},
			GitCryptKey:  "((super.allowed))",
			Branch:       "master",
		},
	}

	assert.Empty(t, secretValidator.Validate(good))
}

func TestFeatureToggles(t *testing.T) {
	bad := manifest.Manifest{
		FeatureToggles: manifest.FeatureToggles{
			"ok",
			"((not.ok))",
			"kehe",
			"((not.ok))",
		},
	}

	errors := secretValidator.Validate(bad)
	assert.Len(t, errors, 2)
	assert.Contains(t, errors, manifest.UnsupportedSecretError("feature_toggles[1]"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("feature_toggles[3]"))
}

func TestRun(t *testing.T) {
	bad := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.Run{
				Type:   "((not.allowed))",
				Name:   "((not.allowed))",
				Script: "./path/to/script ((not.allowed))",
				Docker: manifest.Docker{
					Image:    "((not.allowed))",
					Username: "((super.ok))",
					Password: "((super.ok))",
				},
				Vars: map[string]string{
					"ok":         "((super.ok))",
					"((not.ok))": "blurgh",
				},
				SaveArtifacts: []string{
					"ok",
					"((not.allowed))",
				},
			},
		},
	}

	errors := secretValidator.Validate(bad)
	assert.Len(t, errors, 6)
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].type"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].name"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].script"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].docker.image"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("key tasks[0].vars[((not.ok))]"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].save_artifacts[1]"))

	good := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.Run{
				Type:   "run",
				Name:   "myCoolName",
				Script: "./script",
				Docker: manifest.Docker{
					Image:    "docker-image:tag",
					Username: "((super.secret))",
					Password: "((super.secret))",
				},
				Vars: map[string]string{
					"mySecret":  "((whoop.whoop))",
					"notSecret": "password",
				},

				SaveArtifacts: []string{
					"a",
					"b/c/d",
				},
			},
		},
	}

	assert.Len(t, secretValidator.Validate(good), 0)
}

func TestDockerPush(t *testing.T) {
	bad := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.Run{},
			manifest.DockerPush{
				Type:     "((not.ok))",
				Name:     "((not.ok))",
				Username: "((super.ok))",
				Password: "((super.ok))",
				Image:    "((not.ok))",
				Vars: map[string]string{
					"ok":         "((super.ok))",
					"((not.ok))": "kehe",
				},
			},
		},
	}

	errors := secretValidator.Validate(bad)
	assert.Len(t, errors, 4)
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[1].type"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[1].name"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[1].image"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("key tasks[1].vars[((not.ok))]"))
}

func TestDockerCompose(t *testing.T) {
	bad := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.DockerCompose{
				Type:    "((not.ok))",
				Name:    "((not.ok))",
				Command: "someCommand ((not.ok))",
				Vars: map[string]string{
					"ok":         "((super.secret))",
					"((not.ok))": "blurgh",
				},
				Service: "((not.ok))",
				SaveArtifacts: []string{
					"ok",
					"((not.ok))",
				},
			},
		},
	}

	errors := secretValidator.Validate(bad)
	assert.Len(t, errors, 6)
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].type"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].name"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].command"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("key tasks[0].vars[((not.ok))]"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].service"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].save_artifacts[1]"))
}

func TestDeployCF(t *testing.T) {
	bad := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.DeployCF{
				Type:       "((not.ok))",
				Name:       "((not.ok))",
				API:        "((super.ok))",
				Space:      "((super.ok))",
				Org:        "((super.ok))",
				Username:   "((super.ok))",
				Password:   "((super.ok))",
				Manifest:   "((not.ok))",
				TestDomain: "((super.ok))",
				Vars: map[string]string{
					"ok":         "((super.ok))",
					"((not.ok))": "blurgh",
				},
				DeployArtifact: "((not.ok))",
				Timeout:        "((not.ok))",
			},
		},
	}

	errors := secretValidator.Validate(bad)
	assert.Len(t, errors, 6)
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].type"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].name"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].manifest"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("key tasks[0].vars[((not.ok))]"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].deploy_artifact"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].timeout"))

	badPrePromote := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.DeployCF{
				PrePromote: manifest.TaskList{
					manifest.Run{
						Name: "((not.ok))",
					},
					manifest.DockerCompose{
						Type: "((not.ok))",
					},
				},
			},
		},
	}

	errors = secretValidator.Validate(badPrePromote)
	assert.Len(t, errors, 2)
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].pre_promote[0].name"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].pre_promote[1].type"))
}

func TestConsumerIntegrationTest(t *testing.T) {
	bad := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.ConsumerIntegrationTest{
				Type:                 "((not.ok))",
				Name:                 "((not.ok))",
				Consumer:             "((not.ok))",
				ConsumerHost:         "((not.ok))",
				ProviderHost:         "((not.ok))",
				Script:               "./script ((not.ok))",
				DockerComposeService: "((not.ok))",
				Vars: map[string]string{
					"ok":         "((super.secret))",
					"((not.ok))": "blah",
				},
			},
		},
	}

	errors := secretValidator.Validate(bad)
	assert.Len(t, errors, 8)
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].type"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].name"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].consumer"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].consumer_host"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].provider_host"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].script"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].docker_compose_service"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("key tasks[0].vars[((not.ok))]"))
}

func TestDeployMLZip(t *testing.T) {
	bad := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.DeployMLZip{
				Type:       "((not.ok))",
				Name:       "((not.ok))",
				DeployZip:  "((not.ok))",
				AppName:    "((not.ok))",
				AppVersion: "((not.ok))",
				Targets: []string{
					"((super.ok))",
					"hostBlah.com",
				},
			},
		},
	}
	errors := secretValidator.Validate(bad)
	assert.Len(t, errors, 5)

	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].type"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].name"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].deploy_zip"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].app_name"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].app_version"))
}

func TestDeployMLModules(t *testing.T) {
	bad := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.DeployMLModules{
				Type:             "((not.ok))",
				Name:             "((not.ok))",
				MLModulesVersion: "((not.ok))",
				AppName:          "((not.ok))",
				AppVersion:       "((not.ok))",
				Targets: []string{
					"((super.ok))",
					"hostBlah.com",
				},
			},
		},
	}
	errors := secretValidator.Validate(bad)
	assert.Len(t, errors, 5)

	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].type"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].name"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].ml_modules_version"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].app_name"))
	assert.Contains(t, errors, manifest.UnsupportedSecretError("tasks[0].app_version"))
}

func TestBadKeys(t *testing.T) {
	bad := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.Run{
				Docker: manifest.Docker{
					Password: "((a))",
				},
				Vars: map[string]string{
					"secret": "((a.b.c))",
				},
			},
			manifest.DeployCF{
				API: "((this_is_a_invalid$secret.@with_special_chars))",
				PrePromote: manifest.TaskList{
					manifest.DockerCompose{
						Vars: map[string]string{
							"SuperSecret": "((this_is_a_invalid$secret.@with_special_chars))",
						},
					},
				},
			},
		},
	}

	errors := secretValidator.Validate(bad)
	assert.Len(t, errors, 4)
	assert.Contains(t, errors, manifest.InvalidSecretError("((a))", "tasks[0].docker.password"))
	assert.Contains(t, errors, manifest.InvalidSecretError("((a.b.c))", "tasks[0].vars[secret]"))
	assert.Contains(t, errors, manifest.InvalidSecretError("((this_is_a_invalid$secret.@with_special_chars))", "tasks[1].api"))
	assert.Contains(t, errors, manifest.InvalidSecretError("((this_is_a_invalid$secret.@with_special_chars))", "tasks[1].pre_promote[0].vars[SuperSecret]"))
}

func TestArtifactConfig(t *testing.T) {
	man := manifest.Manifest{
		ArtifactConfig: manifest.ArtifactConfig{
			Bucket:  "((superSecret.bucket))",
			JsonKey: "((superSecret.JsonKey))",
		},
	}

	errs := secretValidator.Validate(man)
	assert.Len(t, errs, 0)
}
