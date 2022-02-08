package databuilder

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
)

func resolveDependencies(mapping map[string]*builder, initData ...string) ([]*builder, error) {
	/*
	 * dependency resolution is NP problem, lets see what we can do
	 */
	outputMap := make(map[string]string)      // mapping between function return and function
	structMap := make(map[string]sets.String) // mapping between output struct and input struct
	for _, v := range mapping {
		outputMap[v.Out] = v.Name
		if _, ok := structMap[v.Out]; !ok {
			structMap[v.Out] = sets.NewString()
		}
		structMap[v.Out].Insert(v.In...)
	}

	readyset := sets.NewString(initData...)
	order := make([]*builder, 0)
	for len(structMap) > 0 {
		blocked := sets.NewString()
		for k, v := range structMap {
			if v.Len() == 0 {
				readyset.Insert(k)
			} else {
				blocked.Insert(v.List()...)
			}
		}
		if readyset.Len() == 0 {
			return make([]*builder, 0), fmt.Errorf("%w: missing fields %s", ErrCouldNotResolveDependency, blocked)
		}
		for _, v := range readyset.List() {
			fn, ok := outputMap[v]
			if !ok {
				// skip already provided fields
				continue
			}
			order = append(order, mapping[fn])
			delete(structMap, v)
		}
		for k, v := range structMap {
			diff := v.Difference(readyset)
			structMap[k] = diff
		}
		readyset = sets.NewString()
	}
	return order, nil
}
