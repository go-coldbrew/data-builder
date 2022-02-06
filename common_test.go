package databuilder

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct1 struct {
}

type TestStruct2 struct {
}

type TestStruct3 struct {
}

type TestInter interface {
}

func DBTestFunc(_ context.Context, _ TestStruct1) (TestStruct2, error) {
	return TestStruct2{}, nil
}

func DBTestFunc2(_ context.Context) (TestStruct1, error) {
	return TestStruct1{}, nil
}

func DBTestFunc3(_ context.Context, _ TestStruct2) (TestStruct1, error) {
	return TestStruct1{}, nil
}

func DBTestFuncInvalid1(_ context.Context, _ int) (TestStruct1, error) {
	return TestStruct1{}, nil
}

func DBTestFuncInvalid2(_ context.Context, _ TestStruct2) TestStruct1 {
	return TestStruct1{}
}

func DBTestFuncInvalid3(_ context.Context, arr ...TestStruct2) (TestStruct1, error) {
	return TestStruct1{}, nil
}

func DBTestFuncInvalid4(_ context.Context, _ *TestStruct2) (TestStruct1, error) {
	return TestStruct1{}, nil
}

func DBTestFuncInvalid5(_ context.Context, _ TestStruct1) (TestStruct1, error) {
	return TestStruct1{}, nil
}

func DBTestFuncInvalid6(_ context.Context, _ TestStruct2) (int, error) {
	return 0, nil
}

func DBTestFuncInvalid7(_ context.Context, _ TestStruct1) (TestStruct1, int) {
	return TestStruct1{}, 0
}

func DBTestFuncInvalid8(_ context.Context, _ TestStruct1) (TestStruct1, TestInter) {
	return TestStruct1{}, nil
}

func testNew(t *testing.T) *db {
	dbuild := New()
	d, ok := dbuild.(*db)
	if !ok {
		assert.Fail(t, "New Should return an object of *db")
	}
	return d
}
