package databuilder

import (
	"context"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// Quiet benchmark-only builder variants. The production fixtures in
// common_test.go call fmt.Println and dominate end-to-end timings, hiding
// the effect we want to measure.

type benchStructIn struct{ Value string }
type benchStructA struct{ Value string }
type benchStructB struct{ Value string }
type benchStructC struct{ Value string }
type benchStructD struct{ Value string }

func benchFuncA(_ context.Context, s benchStructIn) (benchStructA, error) {
	return benchStructA{Value: strings.ReplaceAll(s.Value, "-", "_")}, nil
}

func benchFuncB(_ context.Context, s benchStructA) (benchStructB, error) {
	return benchStructB{Value: s.Value + "B"}, nil
}

func benchFuncC(_ context.Context, s benchStructA) (benchStructC, error) {
	return benchStructC{Value: s.Value + "C"}, nil
}

func benchFuncD(_ context.Context, _ benchStructB, _ benchStructC) (benchStructD, error) {
	return benchStructD{Value: "D"}, nil
}

// uncachedStructName reproduces the pre-caching implementation for apples-to-apples
// comparison in the micro-benchmarks.
func uncachedStructName(t reflect.Type) string {
	return t.PkgPath() + "." + t.Name()
}

// --- struct name resolution ---

func BenchmarkGetStructName_Uncached(b *testing.B) {
	t := reflect.TypeOf(benchStructA{})
	b.ReportAllocs()
	b.ResetTimer()
	var got string
	for i := 0; i < b.N; i++ {
		got = uncachedStructName(t)
	}
	runtime.KeepAlive(got)
}

func BenchmarkCachedStructName_Hit(b *testing.B) {
	t := reflect.TypeOf(benchStructA{})
	_ = cachedStructName(t)
	b.ReportAllocs()
	b.ResetTimer()
	var got string
	for i := 0; i < b.N; i++ {
		got = cachedStructName(t)
	}
	runtime.KeepAlive(got)
}

func BenchmarkCachedStructName_MixedHit(b *testing.B) {
	types := []reflect.Type{
		reflect.TypeOf(benchStructIn{}),
		reflect.TypeOf(benchStructA{}),
		reflect.TypeOf(benchStructB{}),
		reflect.TypeOf(benchStructC{}),
		reflect.TypeOf(benchStructD{}),
	}
	for _, t := range types {
		_ = cachedStructName(t)
	}
	b.ReportAllocs()
	b.ResetTimer()
	var got string
	for i := 0; i < b.N; i++ {
		got = cachedStructName(types[i%len(types)])
	}
	runtime.KeepAlive(got)
}

// --- function PC resolution ---

func BenchmarkFuncForPC_Uncached(b *testing.B) {
	pc := reflect.ValueOf(benchFuncA).Pointer()
	b.ReportAllocs()
	b.ResetTimer()
	var got string
	for i := 0; i < b.N; i++ {
		got = runtime.FuncForPC(pc).Name()
	}
	runtime.KeepAlive(got)
}

func BenchmarkResolveFuncName_Hit(b *testing.B) {
	pc := reflect.ValueOf(benchFuncA).Pointer()
	_ = resolveFuncName(pc)
	b.ReportAllocs()
	b.ResetTimer()
	var got string
	for i := 0; i < b.N; i++ {
		got = resolveFuncName(pc)
	}
	runtime.KeepAlive(got)
}

func BenchmarkResolveFuncName_MixedHit(b *testing.B) {
	pcs := []uintptr{
		reflect.ValueOf(benchFuncA).Pointer(),
		reflect.ValueOf(benchFuncB).Pointer(),
		reflect.ValueOf(benchFuncC).Pointer(),
		reflect.ValueOf(benchFuncD).Pointer(),
	}
	for _, pc := range pcs {
		_ = resolveFuncName(pc)
	}
	b.ReportAllocs()
	b.ResetTimer()
	var got string
	for i := 0; i < b.N; i++ {
		got = resolveFuncName(pcs[i%len(pcs)])
	}
	runtime.KeepAlive(got)
}

// --- registration ---

func BenchmarkAddBuilders(b *testing.B) {
	// Pin cache state to "warm" so this benchmark measures steady-state
	// registration and doesn't drift based on prior benchmark ordering.
	resetCachesForTest()
	warm := New()
	if err := warm.AddBuilders(benchFuncA, benchFuncB, benchFuncC, benchFuncD); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := New()
		if err := d.AddBuilders(benchFuncA, benchFuncB, benchFuncC, benchFuncD); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAddBuilders_ColdCache exercises the worst-case path where the
// caches are purged before every iteration. Not realistic, but it pins the
// ceiling of how much the caches can help registration.
func BenchmarkAddBuilders_ColdCache(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		resetCachesForTest()
		b.StartTimer()
		d := New()
		if err := d.AddBuilders(benchFuncA, benchFuncB, benchFuncC, benchFuncD); err != nil {
			b.Fatal(err)
		}
	}
}

// --- compile ---

func BenchmarkCompile(b *testing.B) {
	d := New()
	if err := d.AddBuilders(benchFuncA, benchFuncB, benchFuncC, benchFuncD); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := d.Compile(benchStructIn{}); err != nil {
			b.Fatal(err)
		}
	}
}

// --- end-to-end execution ---

func newBenchPlan(b *testing.B) Plan {
	b.Helper()
	d := New()
	if err := d.AddBuilders(benchFuncA, benchFuncB, benchFuncC, benchFuncD); err != nil {
		b.Fatal(err)
	}
	plan, err := d.Compile(benchStructIn{})
	if err != nil {
		b.Fatal(err)
	}
	return plan
}

func benchRunParallel(b *testing.B, workers uint) {
	plan := newBenchPlan(b)
	ctx := context.Background()
	in := benchStructIn{Value: "hello-world"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := plan.RunParallel(ctx, workers, in); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRunParallel_Workers1(b *testing.B) { benchRunParallel(b, 1) }
func BenchmarkRunParallel_Workers4(b *testing.B) { benchRunParallel(b, 4) }
func BenchmarkRunParallel_Workers8(b *testing.B) { benchRunParallel(b, 8) }

// --- Result.Get ---

func BenchmarkResultGet(b *testing.B) {
	plan := newBenchPlan(b)
	result, err := plan.RunParallel(context.Background(), 4, benchStructIn{Value: "x"})
	if err != nil {
		b.Fatal(err)
	}
	key := benchStructC{}
	b.ReportAllocs()
	b.ResetTimer()
	var got any
	for i := 0; i < b.N; i++ {
		got = result.Get(key)
	}
	runtime.KeepAlive(got)
}

func BenchmarkResultGet_Parallel(b *testing.B) {
	plan := newBenchPlan(b)
	result, err := plan.RunParallel(context.Background(), 4, benchStructIn{Value: "x"})
	if err != nil {
		b.Fatal(err)
	}
	key := benchStructC{}
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var got any
		for pb.Next() {
			got = result.Get(key)
		}
		runtime.KeepAlive(got)
	})
}
