package pipeline

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestTriggerIntervalNotSet(t *testing.T) {
	man := manifest.Manifest{
		Repo: manifest.Repo{URI: gitDir},
		Tasks: []manifest.Task{
			manifest.Run{Script: "run.sh"},
		},
	}
	config := testPipeline().Render(man)
	resources := config.Resources
	plan := config.Jobs[0].Plan

	//should be 1 resource: git
	assert.Len(t, resources, 1)
	assert.Equal(t, "git", resources[0].Type)

	//should be 2 items in the plan: get git + task
	assert.Len(t, plan, 2)
	assert.Equal(t, gitDir, plan[0].Get)
	assert.True(t, plan[0].Trigger)
	assert.Equal(t, "run run.sh", plan[1].Task)
}

func TestTriggerIntervalSet(t *testing.T) {
	man := manifest.Manifest{
		Repo:            manifest.Repo{URI: gitDir},
		TriggerInterval: "1h",
		Tasks: []manifest.Task{
			manifest.Run{Script: "run.sh"},
		},
	}
	config := testPipeline().Render(man)
	resources := config.Resources
	plan := config.Jobs[0].Plan

	//should be 2 resources - git + timer
	assert.Len(t, resources, 2)
	assert.Equal(t, "git", resources[0].Type)
	assert.Equal(t, "timer 1h", resources[1].Name)
	assert.Equal(t, "time", resources[1].Type)
	assert.Equal(t, "1h", resources[1].Source["interval"])

	//should be 3 things in the plan get git + get timer + task
	assert.Len(t, plan, 2)
	assert.Equal(t, gitDir, (*plan[0].Aggregate)[0].Get)
	assert.True(t, (*plan[0].Aggregate)[1].Trigger)

	assert.Equal(t, "timer 1h", (*plan[0].Aggregate)[1].Get)
	assert.True(t, (*plan[0].Aggregate)[1].Trigger)
	assert.Equal(t, "run run.sh", plan[1].Task)
}
