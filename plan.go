package databuilder

import (
	"context"
	"errors"
	"reflect"

	"k8s.io/apimachinery/pkg/util/sets"
)

type plan struct {
	order    []*builder
	initData sets.String // the initial data required for this plan
}

func (p *plan) Run(ctx context.Context, initData ...interface{}) (Result, error) {
	dataMap := make(map[string]interface{})
	initialData := sets.NewString()
	for _, inter := range initData {
		if inter == nil {
			continue
		}
		t := reflect.TypeOf(inter)
		if t.Kind() != reflect.Struct {
			return nil, ErrInvalidBuilderInput
		}
		name := getStructName(t)
		if initialData.Has(name) {
			return nil, ErrMultipleInitialData
		}
		initialData.Insert(name)
		dataMap[name] = inter
	}
	if p.initData.Difference(initialData).Len() > 0 {
		return nil, ErrInitialDataMissing
	}
	return dataMap, p.run(ctx, dataMap)
}

func (p *plan) run(ctx context.Context, dataMap map[string]interface{}) error {
	for i := range p.order {
		b := p.order[i]
		v := reflect.ValueOf(b.Builder)
		input := make([]reflect.Value, 1)
		ctx = AddResultToCtx(ctx, dataMap) // allow builders to access already built data
		input[0] = reflect.ValueOf(ctx)
		for _, in := range b.In {
			data, ok := dataMap[in]
			if !ok {
				return errors.New("TODO: CRITICAL")
			}
			input = append(input, reflect.ValueOf(data))
		}
		outputs := v.Call(input)
		// we should only ever have two outputs
		// 0-> data, 1-> error
		if !outputs[1].IsNil() {
			// error occured, return it back and stop processing
			return outputs[1].Interface().(error)
		}
		name := getStructName(outputs[0].Type())
		dataMap[name] = outputs[0].Interface()
	}
	return nil
}

func (r Result) Get(obj interface{}) interface{} {
	if obj == nil || r == nil {
		return nil
	}
	t := reflect.TypeOf(obj)
	if t.Kind() != reflect.Struct {
		return nil
	}
	name := getStructName(t)
	if value, ok := r[name]; ok {
		return value
	}
	return nil
}

func newPlan(order []*builder, initData []string) (Plan, error) {
	return &plan{
		order:    order,
		initData: sets.NewString(initData...),
	}, nil
}
