package databuilder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveDependenciesNoError(t *testing.T) {
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
	deps["Name4"] = &builder{
		Name: "Name4",
		In:   []string{"A"},
		Out:  "D",
	}

	order, err := resolveDependencies(deps)
	assert.NoError(t, err)
	assert.Len(t, order, len(deps), "all function should be executed")
}

func TestResolveDependenciesErrorExtra(t *testing.T) {
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
	deps["Name4"] = &builder{
		Name: "Name4",
		In:   []string{"D"},
		Out:  "C",
	}

	_, err := resolveDependencies(deps)
	assert.Error(t, err)
}

func TestResolveDependenciesErrorCircular(t *testing.T) {
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
	deps["Name4"] = &builder{
		Name: "Name4",
		In:   []string{"C"},
		Out:  "A",
	}

	_, err := resolveDependencies(deps)
	assert.Error(t, err)
}
