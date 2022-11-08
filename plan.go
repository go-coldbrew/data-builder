package databuilder

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"sync"

	"github.com/go-coldbrew/tracing"
	graphviz "github.com/goccy/go-graphviz"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	// ErrWTF is the error returned in case we find dependency resolution related errors, please report these
	ErrWTF = errors.New("What a Terrible Failure!, This is likely a bug in dependency resolution, please report this :|")
)

type plan struct {
	order    [][]*builder
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
		for j := range p.order[i] {
			b := p.order[i][j]
			if f.Name == b.Name {
				// same function, lets replace it
				p.order[i][j] = t
				return nil
			}
		}
	}
	return errors.New("builder not found")
}

func (p *plan) Run(ctx context.Context, initData ...interface{}) (Result, error) {
	span, ctx := tracing.NewInternalSpan(ctx, "DBRun")
	defer span.End()
	r, err := p.RunParallel(ctx, 1, initData...)
	if err != nil {
		span.SetError(err)
	}
	return r, err
}

func (p *plan) RunParallel(ctx context.Context, workers uint, initData ...interface{}) (Result, error) {
	span, ctx := tracing.NewInternalSpan(ctx, "DBRunParallel")
	defer span.End()
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
	span.SetTag("workers", workers)
	if p.initData.Difference(initialData).Len() > 0 {
		return nil, span.SetError(ErrInitialDataMissing)
	}
	return dataMap, span.SetError(p.run(ctx, workers, dataMap))
}

type work struct {
	out     chan<- output
	wg      *sync.WaitGroup
	builder *builder
	dataMap map[string]interface{}
}

type output struct {
	outputs []reflect.Value
	builder *builder
	err     error
}

func worker(ctx context.Context, wChan <-chan work) {
	for w := range wChan {
		processWork(ctx, w)
	}
}

func processWork(ctx context.Context, w work) {
	defer w.wg.Done() // ensure we close wait group
	o := output{builder: w.builder}
	fn := reflect.ValueOf(w.builder.Builder)
	args := make([]reflect.Value, 1)
	trace, ctx := tracing.NewInternalSpan(ctx, w.builder.Name)
	defer trace.End()
	args[0] = reflect.ValueOf(ctx) // first arg is context.Context
	for _, in := range w.builder.In {
		data, ok := w.dataMap[in]
		if !ok {
			o.err = ErrWTF
			w.out <- o
			return
		}
		args = append(args, reflect.ValueOf(data))
	}
	o.outputs = fn.Call(args)
	if len(o.outputs) > 1 && !o.outputs[1].IsNil() {
		trace.SetError(o.outputs[1].Interface().(error))
	}
	w.out <- o
}

func doWorkAndGetResult(ctx context.Context, builders []*builder, dataMap map[string]interface{}, wChan chan<- work) error {
	// create a output channel to read results
	outChan := make(chan output, len(builders)+1)
	// create a wait group to wait for all results
	var wg sync.WaitGroup
	ctx = AddResultToCtx(ctx, dataMap) // allow builders to access already built data
	for j := range builders {
		b := builders[j]
		if _, ok := dataMap[b.Out]; ok {
			// do not run the builder if the data already exists
			continue
		}
		// build work
		w := work{}
		w.builder = b
		w.wg = &wg
		w.dataMap = dataMap
		w.out = outChan
		wg.Add(1)  // increment count
		wChan <- w // send work to be done by workers
	}
	wg.Wait() // wait for work to be processed
	close(outChan)
	for o := range outChan {
		if o.err != nil {
			return o.err
		}
		outputs := o.outputs
		// we should only ever have two outputs
		// 0-> data, 1-> error
		if !outputs[1].IsNil() {
			// error occured, return it back and stop processing
			return outputs[1].Interface().(error)
		}
		// add result
		name := getStructName(outputs[0].Type())
		dataMap[name] = outputs[0].Interface()
	}
	return nil
}

func (p *plan) run(ctx context.Context, workers uint, dataMap map[string]interface{}) error {
	if workers == 0 {
		workers = 1
	}

	// create a work channel and start workers
	wChan := make(chan work, 0)
	defer close(wChan)
	for i := uint(0); i < workers; i++ {
		go worker(ctx, wChan)
	}

	for i := range p.order {
		err := doWorkAndGetResult(ctx, p.order[i], dataMap, wChan)
		if err != nil {
			return err
		}
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
		for j := range p.order[i] {
			b := p.order[i][j]
			fn, err := graph.CreateNode(b.Name + " [" + strconv.Itoa(i) + "]") // here [] denotes order
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
	}
	return g.RenderFilename(graph, graphviz.Format(format), file)
}

func newPlan(order [][]*builder, initData []string) (Plan, error) {
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

// MaxPlanParallelism return the maximum number of buildes that can be exsecuted parallely
// for a given plan
//
// this number does not take into account if the builder are cpu intensive or netwrok intensive
// it may not be benificial to run builders at max parallelism if they are cpu intensive
func MaxPlanParallelism(pl Plan) (uint, error) {
	p, ok := pl.(*plan)
	if !ok {
		return 0, errors.New("could not find plan created by data-builder")
	}
	max := 1
	for _, order := range p.order {
		if len(order) > max {
			max = len(order)
		}
	}
	return uint(max), nil
}
