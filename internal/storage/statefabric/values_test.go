package statefabric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStepsFromCondition_FiltersPipelineSteps(t *testing.T) {
	steps := stepsFromCondition(map[string]interface{}{
		"steps": []interface{}{
			map[string]interface{}{"name": "step1", "type": "function"},
		},
	})
	assert.Len(t, steps, 1)

	empty := stepsFromCondition(map[string]interface{}{"description": "pipeline only"})
	assert.Nil(t, empty)
}

func TestStrPtr(t *testing.T) {
	assert.Nil(t, strPtr(""))
	assert.Equal(t, "abc", *strPtr("abc"))
}
