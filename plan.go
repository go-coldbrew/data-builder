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

// ErrWTF is the error returned in case we find dependency resolution related errors, please report these
var ErrWTF = errors.New("What a Terrible Failure!, This is likely a bug in dependency resolution, please report this :|")

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
	return r, span.SetError(err)
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
	span, ctx := tracing.NewInternalSpan(ctx, w.builder.Name)
	defer span.End()
	o := output{builder: w.builder}
	defer func() {
		// recover from panic and set error
		if r := recover(); r != nil {
			o.err = span.SetError(errors.New("panic in builder: " + w.builder.Name))
			w.out <- o
		}
	}()
	fn := reflect.ValueOf(w.builder.Builder)
	// allow builders to access already built data
	ctx = AddResultToCtx(ctx, w.dataMap)
	args := make([]reflect.Value, 1)
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
		span.SetError(o.outputs[1].Interface().(error)) // nolint: errcheck
	}
	w.out <- o
}

func doWorkAndGetResult(ctx context.Context, builders []*builder, dataMap map[string]interface{}, wChan chan<- work) error {
	// create a output channel to read results
	outChan := make(chan output, len(builders)+1)
	// create a wait group to wait for all results
	var wg sync.WaitGroup
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
	errs := make([]error, 0)
	for o := range outChan {
		if o.err != nil {
			// error occured, return it back and stop processing
			return o.err
		}
		outputs := o.outputs
		// we should only ever have two outputs
		// 0-> data, 1-> error
		if !outputs[1].IsNil() {
			// error occured, add it to the list of errors and continue processing
			errs = append(errs, outputs[1].Interface().(error))
			continue
		}
		// add result
		name := getStructName(outputs[0].Type())
		dataMap[name] = outputs[0].Interface()
	}
	if len(errs) > 0 {
		// we only return the first error
		// TODO enhance error handling
		return errs[0]
	}
	return nil
}

func (p *plan) run(ctx context.Context, workers uint, dataMap map[string]interface{}) error {
	if workers == 0 {
		workers = 1
	}

	// create a work channel and start workers
	wChan := make(chan work)
	defer close(wChan)
	for i := uint(0); i < workers; i++ {
		go worker(ctx, wChan)
	}

	errs := make([]error, 0)
	for i := range p.order {
		err := doWorkAndGetResult(ctx, p.order[i], dataMap, wChan)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		// we only return the first error
		// TODO enhance error handling
		return errs[0]
	}
	return nil
}

// Result.Get returns the value of the struct from the result
// if the struct is not found in the result, nil is returned
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

// BuildGraph builds a graphviz graph of the dependency graph of the plan and writes it to the file specified.
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
			_, err = graph.CreateEdge("Out", fn, out)
			if err != nil {
				return err
			}
			for _, in := range b.In {
				in, err := graph.CreateNode(in)
				if err != nil {
					return err
				}
				in = in.SetFontColor(STRUCTCOLOR)
				_, err = graph.CreateEdge("In", in, fn)
				if err != nil {
					return err
				}
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
