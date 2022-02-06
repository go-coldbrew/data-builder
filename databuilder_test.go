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
	var intVal int = 1
	assert.Error(t, IsValidBuilder(intVal), "Non function values should NOT be valid")
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
}

func TestPlan(t *testing.T) {
}
