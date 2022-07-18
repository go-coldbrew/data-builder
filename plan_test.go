package databuilder

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestPlanRun(t *testing.T) {
	const VALUE = "9F0D8E07-6C46-48B7-983C-5C309C042CC6"

	d := testNew(t)
	err := d.AddBuilders(DBTestFunc, DBTestFunc4)
	assert.NoError(t, err)
	executionPlan, err := d.Compile(TestStruct1{})
	assert.NotNil(t, executionPlan)
	assert.NoError(t, err)

	ctx := context.Background()

	_, err = executionPlan.Run(ctx, 1)
	assert.Error(t, err, "initial data needs to be struct")
	_, err = executionPlan.Run(ctx, TestStruct1{}, TestStruct1{})
	assert.Error(t, err, "multiple instance of same data should not be provided")
	_, err = executionPlan.Run(ctx)
	assert.Error(t, err, "missing starting data should error out")

	result, err := executionPlan.Run(ctx,
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
	goleak.VerifyNone(t)
}

func ExamplePlan() {
	b := New()
	err := b.AddBuilders(DBTestFunc, DBTestFunc4)
	fmt.Println(err == nil)
	ep, err := b.Compile(TestStruct1{})
	fmt.Println(err == nil)

	_, err = ep.Run(context.Background(), TestStruct1{})
	fmt.Println(err == nil)

	err = ep.Replace(context.Background(), DBTestFunc, DBTestFunc5)
	fmt.Println(err == nil)
	_, err = ep.Run(context.Background(), TestStruct1{})
	fmt.Println(err == nil)

	// Output:
	// true
	// true
	// CALLED DBTestFunc
	// CALLED DBTestFunc4
	// true
	// true
	// CALLED DBTestFunc5
	// CALLED DBTestFunc4
	// true
}
