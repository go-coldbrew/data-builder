package databuilder

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
)

func resolveDependencies(mapping map[string]*builder) error {
	outputMap := make(map[string]string)     // mapping between function return and function
	inputMap := make(map[string]sets.String) // mapping between function input and function
	for _, v := range mapping {
		outputMap[v.Out] = v.Name
		for _, out := range v.In {
			if _, ok := inputMap[out]; !ok {
				inputMap[out] = sets.NewString()
			}
			inputMap[out].Insert(v.Name)
		}
	}
	fmt.Println("out->", outputMap)
	fmt.Println("in->", inputMap)
	return nil
}
