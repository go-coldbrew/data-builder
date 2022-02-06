package databuilder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	assert.Error(t, IsValidBuilder(DBTestFuncInvalid5), "DBTestFuncInvalid5 should NOT be valid")
	assert.Error(t, IsValidBuilder(DBTestFuncInvalid6), "DBTestFuncInvalid6 should NOT be valid")
	assert.Error(t, IsValidBuilder(DBTestFuncInvalid7), "DBTestFuncInvalid7 should NOT be valid")
	assert.Error(t, IsValidBuilder(DBTestFuncInvalid8), "DBTestFuncInvalid8 should NOT be valid")
	var intVal int = 1
	assert.Error(t, IsValidBuilder(intVal), "Non function values should NOT be valid")
	assert.Error(t, IsValidBuilder(TestStruct2{}), "Non function values should NOT be valid")

}

func TestAddbuilder(t *testing.T) {
	dbuild := New()
	d, ok := dbuild.(*db)
	if !ok {
		assert.Fail(t, "New Should return an object of *db")
	}
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
	dbuild := New()
	d, ok := dbuild.(*db)
	if !ok {
		assert.Fail(t, "New Should return an object of *db")
	}
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
	dbuild := New()
	d, ok := dbuild.(*db)
	if !ok {
		assert.Fail(t, "New Should return an object of *db")
	}
	err := d.AddBuilders(DBTestFunc, DBTestFunc3)
	assert.NoError(t, err)
	_, err = d.Compile()
	assert.Error(t, err, "cyclic dependency should return an error")
}

func TestPlan(t *testing.T) {
}
