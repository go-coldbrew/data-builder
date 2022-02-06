package databuilder

import (
	"errors"
	"reflect"

	"k8s.io/apimachinery/pkg/util/sets"
)

type plan struct {
	order    []*builder
	initData sets.String            // the initial data required for this plan
	dataMap  map[string]interface{} // map of all data
}

func (p *plan) Run(initData ...interface{}) (Data, error) {
	initialialData := sets.NewString()
	for _, inter := range initData {
		if inter == nil {
			continue
		}
		t := reflect.TypeOf(inter)
		if t.Kind() != reflect.Struct {
			return nil, errors.New("invalid initial data, needs to be struct")
		}
		name := getStructName(t)
		if initialialData.Has(name) {
			return nil, errors.New("initial data provided twice")
		}
		initialialData.Insert(name)
	}
	return nil, nil

}

func newPlan(order []*builder, initData []string) (Plan, error) {
	return &plan{
		order:    order,
		initData: sets.NewString(initData...),
	}, nil
}
