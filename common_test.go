package databuilder

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct1 struct {
	Value string
}

type TestStruct2 struct {
	Value string
}

type TestStruct3 struct {
	Value string
}

type TestInter interface {
}

func DBTestFunc(_ context.Context, s TestStruct1) (TestStruct2, error) {
	fmt.Println("CALLED DBTestFunc")
	return TestStruct2{
		Value: strings.ReplaceAll(s.Value, "-", "_"),
	}, nil
}

func DBTestFunc2(_ context.Context) (TestStruct1, error) {
	fmt.Println("CALLED DBTestFunc2")
	return TestStruct1{
		Value: "ABCD",
	}, nil
}

func DBTestFunc3(_ context.Context, s TestStruct2) (TestStruct1, error) {
	fmt.Println("CALLED DBTestFunc3")
	return TestStruct1{
		Value: s.Value,
	}, nil
}

func DBTestFunc4(_ context.Context, s TestStruct1) (TestStruct3, error) {
	fmt.Println("CALLED DBTestFunc4")
	return TestStruct3{
		Value: s.Value,
	}, nil
}

func DBTestFunc5(_ context.Context, s TestStruct1) (TestStruct2, error) {
	fmt.Println("CALLED DBTestFunc5")
	return TestStruct2{
		Value: strings.ReplaceAll(s.Value, "-", "--"),
	}, nil
}

func DBTestFunc6(_ context.Context, s TestStruct1) (TestStruct3, error) {
	fmt.Println("CALLED DBTestFunc6")
	return TestStruct3{
		Value: s.Value,
	}, fmt.Errorf("DBTestFunc6 encountered an error")
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
