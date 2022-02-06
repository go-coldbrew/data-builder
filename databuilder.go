package databuilder

import (
	"context"
	"reflect"
	"runtime"

	"k8s.io/apimachinery/pkg/util/sets"
)

/*
 * Before going throught this code please read - https://go.dev/blog/laws-of-reflection
 */

type builder struct {
	Builder interface{}
	In      []string
	Out     string
	Name    string
}

type db struct {
	builders map[string]*builder
	outSet   sets.String
}

func (d *db) AddBuilders(builders ...interface{}) error {
	// initialize
	if d.builders == nil {
		d.builders = make(map[string]*builder)
	}
	if d.outSet == nil {
		d.outSet = sets.NewString()
	}

	// go through all builders and add them
	for i := range builders {
		b := builders[i]
		if b == nil {
			return ErrInvalidBuilder
		}
		if err := d.add(b); err != nil {
			return err
		}
	}
	return nil
}

func (d *db) add(bldr interface{}) error {
	if err := IsValidBuilder(bldr); err != nil {
		return err
	}

	t := reflect.TypeOf(bldr)
	out := getStructName(t.Out(0))
	name := getFuncName(bldr)

	// check for name
	if _, ok := d.builders[name]; ok {
		return nil
	}

	//check for outSet
	if d.outSet.Has(out) {
		return ErrMultipleBuilderSameOutput
	}

	b := &builder{
		Out:     out,
		Builder: bldr,
		Name:    name,
	}
	// first in context.Context so we start from second
	for i := 1; i < t.NumIn(); i++ {
		b.In = append(b.In, getStructName(t.In(i)))
	}
	d.builders[name] = b
	d.outSet.Insert(out)
	return nil
}

func (d *db) Compile(init ...interface{}) (Plan, error) {
	initialialData := make([]string, 0, len(init))
	for _, inter := range init {
		if inter == nil {
			continue
		}
		t := reflect.TypeOf(inter)
		if t.Kind() != reflect.Struct {
			return nil, ErrInvalidBuilderInput
		}
		initialialData = append(initialialData, getStructName(t))
	}

	order, err := resolveDependencies(d.builders, initialialData...)
	if err != nil {
		return nil, err
	}
	return newPlan(order, initialialData)
}

// IsValidBuilder checks if the given function is valid or not
func IsValidBuilder(builder interface{}) error {
	t := reflect.TypeOf(builder)
	if t.Kind() != reflect.Func {
		// Input can only be a function
		return ErrInvalidBuilderKind
	}
	if t.NumOut() != 2 {
		// should return a struct and an error
		return ErrInvalidBuilderNumOutput
	}
	if t.Out(0).Kind() != reflect.Struct {
		// first return argument should always be a struct
		return ErrInvalidBuilderFirstOutput
	}
	if t.Out(1).Kind() != reflect.Interface {
		// second return argument should always be an Interface
		return ErrInvalidBuilderSecondOutput
	}
	if !t.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		// second return argument should always be an error
		return ErrInvalidBuilderSecondOutput
	}
	if t.NumIn() > 0 {
		// first input should always be context.Context
		if t.In(0).Kind() != reflect.Interface {
			return ErrInvalidBuilderMissingContext
		}
		if !t.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
			return ErrInvalidBuilderMissingContext
		}
		// other inputs should all be structs
		for i := 1; i < t.NumIn(); i++ {
			if t.In(i).Kind() != reflect.Struct {
				// checks for vardic functions as well
				return ErrInvalidBuilderInput
			}
			if getStructName(t.In(i)) == getStructName(t.Out(0)) {
				return ErrSameInputAsOutput
			}
		}
	} else {
		return ErrInvalidBuilderMissingContext
	}

	return nil
}

func getFuncName(bldr interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(bldr).Pointer()).Name()
}

func getStructName(t reflect.Type) string {
	return t.PkgPath() + "." + t.Name()
}

// New Creates a new DataBuilder
func New() DataBuilder {
	return &db{}
}
