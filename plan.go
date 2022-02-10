package databuilder

import (
	"context"
	"errors"
	"reflect"

	graphviz "github.com/goccy/go-graphviz"
	"k8s.io/apimachinery/pkg/util/sets"
)

type plan struct {
	order    []*builder
	initData sets.String // the initial data required for this plan
}

func (p *plan) Replace(ctx context.Context, from interface{}, to interface{}) error {
	f, err := getBuilder(from)
	if err != nil {
		return err
	}

	t, err := getBuilder(to)
	if err != nil {
		return err
	}

	if f.Name == t.Name {
		// same function, do nothing
		return nil
	}

	if f.Out != t.Out {
		return errors.New("both builders should have the same output")
	}

	input := sets.NewString(f.In...)
	if !input.IsSuperset(sets.NewString(t.In...)) {
		return errors.New("replace can NOT introduce dependencies, please compile a new plan")
	}

	for i := range p.order {
		b := p.order[i]
		if f.Name == b.Name {
			// same function, lets replace it
			p.order[i] = t
			return nil
		}
	}
	return errors.New("builder not found")
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
		if _, ok := dataMap[b.Out]; ok {
			// do not run the builder if the data already exists
			continue
		}
		v := reflect.ValueOf(b.Builder)
		input := make([]reflect.Value, 1)
		ctx = AddResultToCtx(ctx, dataMap) // allow builders to access already built data
		input[0] = reflect.ValueOf(ctx)
		for _, in := range b.In {
			data, ok := dataMap[in]
			if !ok {
				return errors.New("What a Terrible Failure!, This is likely a bug in dependency resolution, please report this :|")
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

func (p plan) BuildGraph(format, file string) error {
	const (
		FNCOLOR     = "red"
		STRUCTCOLOR = "blue"
	)

	g := graphviz.New()
	graph, err := g.Graph(graphviz.Name("Dependency Graph"))
	if err != nil {
		return err
	}
	for i := range p.order {
		b := p.order[i]
		fn, err := graph.CreateNode(b.Name)
		if err != nil {
			return err
		}
		fn = fn.SetFontColor(FNCOLOR)
		out, err := graph.CreateNode(b.Out)
		if err != nil {
			return err
		}
		out = out.SetFontColor(STRUCTCOLOR)
		graph.CreateEdge("Out", fn, out)
		for _, in := range b.In {
			in, err := graph.CreateNode(in)
			if err != nil {
				return err
			}
			in = in.SetFontColor(STRUCTCOLOR)
			graph.CreateEdge("In", in, fn)
		}
	}
	return g.RenderFilename(graph, graphviz.Format(format), file)
}

func newPlan(order []*builder, initData []string) (Plan, error) {
	return &plan{
		order:    order,
		initData: sets.NewString(initData...),
	}, nil
}

// BuildGraph helps understand the execution plan, it renders the plan in the given format
// please note we depend on graphviz, please ensure you have graphviz installed
func BuildGraph(executionPlan Plan, format, file string) error {
	if p, ok := executionPlan.(*plan); ok {
		return p.BuildGraph(format, file)
	}
	return errors.New("could not find graph builder")
}
