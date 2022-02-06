package databuilder

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlanRun(t *testing.T) {
	const VALUE = "9F0D8E07-6C46-48B7-983C-5C309C042CC6"

	d := testNew(t)
	err := d.AddBuilders(DBTestFunc, DBTestFunc4)
	assert.NoError(t, err)
	plan, err := d.Compile(TestStruct1{})
	assert.NotNil(t, plan)
	assert.NoError(t, err)

	ctx := context.Background()

	_, err = plan.Run(ctx, 1)
	assert.Error(t, err, "initial data needs to be struct")
	_, err = plan.Run(ctx, TestStruct1{}, TestStruct1{})
	assert.Error(t, err, "multiple instance of same data should not be provided")
	_, err = plan.Run(ctx)
	assert.Error(t, err, "missing starting data should error out")

	result, err := plan.Run(ctx,
		nil,
		TestStruct1{
			Value: VALUE,
		},
	) // nil values should be ignored
	assert.NoError(t, err)
	var t3 TestStruct3
	data := result.Get(t3)
	assert.NotNil(t, data)
	ts3, ok := data.(TestStruct3)
	assert.True(t, ok)
	assert.Equal(t, VALUE, ts3.Value)

	var t2 TestStruct2
	data = result.Get(t2)
	assert.NotNil(t, data)
	ts2, ok := data.(TestStruct2)
	assert.True(t, ok)
	assert.Equal(t, strings.ReplaceAll(VALUE, "-", "_"), ts2.Value)

}
