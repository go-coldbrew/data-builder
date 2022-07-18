package databuilder

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
	count := 0
	for i := range order {
		for range order[i] {
			count += 1
		}
	}
	assert.Equal(t, count, len(deps), "all function should present")
}

func TestResolveParallelDependenciesNoError(t *testing.T) {
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
	deps["Name5"] = &builder{
		Name: "Name5",
		In:   []string{"B"},
		Out:  "F",
	}

	order, err := resolveDependencies(deps)
	assert.NoError(t, err)

	names := make([][]string, 0)
	count := 0
	for i := range order {
		n := make([]string, 0)
		for j := range order[i] {
			count += 1
			n = append(n, order[i][j].Name)
		}
		names = append(names, n)
	}
	assert.Equal(t, count, len(deps), "all function should be executed")
	exp := [][]string{{"Name3"}, {"Name2", "Name5"}, {"Name1", "Name4"}}
	assert.True(t,
		cmp.Equal(names, exp, cmpopts.SortSlices(func(s1, s2 string) bool { return s1 > s2 })),
		"Order should match"+cmp.Diff(order, exp, cmpopts.SortSlices(func(s1, s2 string) bool { return s1 > s2 })),
	)
}

func TestResolveDependenciesDisjointGraphsNoError(t *testing.T) {
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
		In:   []string{},
		Out:  "Y",
	}
	deps["Name5"] = &builder{
		Name: "Name5",
		In:   []string{"Y"},
		Out:  "Z",
	}

	order, err := resolveDependencies(deps)
	assert.NoError(t, err)

	names := make([][]string, 0)
	count := 0
	for i := range order {
		n := make([]string, 0)
		for j := range order[i] {
			count += 1
			n = append(n, order[i][j].Name)
		}
		names = append(names, n)
	}
	assert.Equal(t, count, len(deps), "all function should be executed")
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
