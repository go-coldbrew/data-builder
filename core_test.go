package databuilder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct1 struct {
}

type TestStruct2 struct {
}

func DBTestFunc(_ TestStruct1) (TestStruct2, error) {
	return TestStruct2{}, nil
}

func DBTestFunc2() (TestStruct1, error) {
	return TestStruct1{}, nil
}

func DBTestFuncInvalid1(_ int) (TestStruct1, error) {
	return TestStruct1{}, nil
}

func DBTestFuncInvalid2(_ TestStruct2) TestStruct1 {
	return TestStruct1{}
}

func DBTestFuncInvalid3(...TestStruct2) (TestStruct1, error) {
	return TestStruct1{}, nil
}

func DBTestFuncInvalid4(_ *TestStruct2) (TestStruct1, error) {
	return TestStruct1{}, nil
}

func TestNew(t *testing.T) {
	dbuild := New()
	d, ok := dbuild.(*db)
	if !ok {
		assert.Fail(t, "New Should return an object of *db")
	}
	err := d.AddBuilders(DBTestFunc, DBTestFunc2)
	assert.NoError(t, err)
}

func TestValidBuilder(t *testing.T) {
	assert.NoError(t, IsValidBuilder(DBTestFunc), "DBTestFunc should be valid")
	assert.NoError(t, IsValidBuilder(DBTestFunc2), "DBTestFunc2 should be valid")
	assert.Error(t, IsValidBuilder(DBTestFuncInvalid1), "DBTestFuncInvalid1 should NOT be valid")
	assert.Error(t, IsValidBuilder(DBTestFuncInvalid2), "DBTestFuncInvalid2 should NOT be valid")
	assert.Error(t, IsValidBuilder(DBTestFuncInvalid3), "DBTestFuncInvalid3 should NOT be valid")
	assert.Error(t, IsValidBuilder(DBTestFuncInvalid4), "DBTestFuncInvalid4 should NOT be valid")
	var intVal int = 1
	assert.Error(t, IsValidBuilder(intVal), "Non function values should NOT be valid")
}

//func TestCompile(t *testing.T) {
//dbuild := New()
//d, ok := dbuild.(*db)
//if !ok {
//assert.Fail(t, "New Should return an object of *db")
//}
//err := d.AddBuilders(DBTestFunc, DBTestFunc2)
//assert.NoError(t, err)
//_, err = d.Compile()
//assert.NoError(t, err)
//}

func TestResolveDependencies(t *testing.T) {
	deps := make(map[string]*builder)
	deps["Name1"] = &builder{
		Name: "Name1",
		In:   []string{"A", "B"},
		Out:  "C",
	}
	deps["Name2"] = &builder{
		Name: "Name2",
		In:   []string{"B"},
		Out:  "A",
	}
	deps["Name3"] = &builder{
		Name: "Name3",
		In:   []string{},
		Out:  "B",
	}

	resolveDependencies(deps)
}
