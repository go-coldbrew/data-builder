package databuilder

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	testNew(t)
}

func TestValidBuilder(t *testing.T) {
	assert.NoError(t, IsValidBuilder(DBTestFunc), "DBTestFunc should be valid")
	assert.NoError(t, IsValidBuilder(DBTestFunc2), "DBTestFunc2 should be valid")
	assert.Error(t, IsValidBuilder(DBTestFuncInvalid1), "DBTestFuncInvalid1 should NOT be valid")
	assert.Error(t, IsValidBuilder(DBTestFuncInvalid2), "DBTestFuncInvalid2 should NOT be valid")
	assert.Error(t, IsValidBuilder(DBTestFuncInvalid3), "DBTestFuncInvalid3 should NOT be valid")
	assert.Error(t, IsValidBuilder(DBTestFuncInvalid4), "DBTestFuncInvalid4 should NOT be valid")
	assert.Error(t, IsValidBuilder(DBTestFuncInvalid5), "DBTestFuncInvalid5 should NOT be valid")
	assert.Error(t, IsValidBuilder(DBTestFuncInvalid6), "DBTestFuncInvalid6 should NOT be valid")
	assert.Error(t, IsValidBuilder(DBTestFuncInvalid7), "DBTestFuncInvalid7 should NOT be valid")
	assert.Error(t, IsValidBuilder(DBTestFuncInvalid8), "DBTestFuncInvalid8 should NOT be valid")
	var intVal int = 1
	assert.Error(t, IsValidBuilder(intVal), "Non function values should NOT be valid")
	assert.Error(t, IsValidBuilder(TestStruct2{}), "Non function values should NOT be valid")

}

func TestAddbuilder(t *testing.T) {
	d := testNew(t)
	err := d.AddBuilders(DBTestFunc, DBTestFunc2)
	assert.NoError(t, err)
	err = d.AddBuilders(DBTestFunc)
	assert.NoError(t, err, "Adding the same builder multiple times should not result in error")
	err = d.AddBuilders(DBTestFunc3)
	assert.Error(t, err)
	err = d.AddBuilders(DBTestFuncInvalid6)
	assert.Error(t, err)
	err = d.AddBuilders(nil)
	assert.Error(t, err)

}

func TestCompile(t *testing.T) {
	d := testNew(t)
	err := d.AddBuilders(DBTestFunc, DBTestFunc2)
	assert.NoError(t, err)
	_, err = d.Compile()
	assert.NoError(t, err)
	_, err = d.Compile(TestStruct2{}, nil)
	assert.NoError(t, err)
	_, err = d.Compile(TestStruct2{}, 0)
	assert.Error(t, err)
}

func TestCompileCyclic(t *testing.T) {
	d := testNew(t)
	err := d.AddBuilders(DBTestFunc, DBTestFunc3)
	assert.NoError(t, err)
	_, err = d.Compile()
	assert.Error(t, err, "cyclic dependency should return an error")
}

func TestTypedNilBuilderRejected(t *testing.T) {
	var nilFunc func(context.Context) (TestStruct1, error)
	d := testNew(t)
	err := d.AddBuilders(nilFunc)
	assert.Error(t, err, "typed-nil func should be rejected")
	assert.ErrorIs(t, err, ErrInvalidBuilder)

	// IsValidBuilder should also reject typed-nil builders directly
	err = IsValidBuilder(nilFunc)
	assert.Error(t, err, "IsValidBuilder should reject typed-nil func")
	assert.ErrorIs(t, err, ErrInvalidBuilder)
}

func TestContextCancellation(t *testing.T) {
	d := testNew(t)
	err := d.AddBuilders(DBTestFunc, DBTestFunc4)
	assert.NoError(t, err)
	plan, err := d.Compile(TestStruct1{})
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err = plan.Run(ctx, TestStruct1{Value: "test"})
	assert.Error(t, err, "cancelled context should return an error")
	assert.ErrorIs(t, err, context.Canceled)
}

func TestJoinErrors(t *testing.T) {
	// Single error should be returned unwrapped
	sentinel := ErrWTF
	err := joinErrors([]error{sentinel})
	assert.Equal(t, sentinel, err, "single error should be returned as-is, not wrapped")

	// No errors
	err = joinErrors(nil)
	assert.NoError(t, err)

	// Multiple errors
	err = joinErrors([]error{ErrWTF, ErrInvalidBuilder})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWTF)
	assert.ErrorIs(t, err, ErrInvalidBuilder)
}
