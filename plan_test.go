package databuilder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlanRun(t *testing.T) {
	d := testNew(t)
	err := d.AddBuilders(DBTestFunc, DBTestFunc2)
	assert.NoError(t, err)
	plan, err := d.Compile()
	assert.NotNil(t, plan)
	assert.NoError(t, err)

	_, err = plan.Run(nil)
	assert.NoError(t, err)
	_, err = plan.Run(1)
	assert.Error(t, err, "initial data needs to be struct")

}
