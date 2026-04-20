// Package databuilder compiles a set of builder functions into an execution
// plan with automatic dependency resolution, then runs them sequentially or
// in parallel.
//
// # Builder functions
//
// A builder is a plain Go function whose signature encodes its inputs and
// output as types:
//
//	func(ctx context.Context, in1 StructA, in2 StructB) (StructC, error)
//
// Rules enforced by [IsValidBuilder]:
//
//   - The first parameter must be context.Context.
//   - All remaining parameters must be concrete struct values (no pointers,
//     no variadics, no primitives).
//   - The function must return exactly two values: a concrete struct and an
//     error.
//   - Two registered builders cannot produce the same output struct.
//   - A builder cannot take its own output type as input.
//
// Types are identified by their fully qualified "pkgpath.TypeName", so the
// dependency graph is built entirely from ordinary Go type information.
//
// # Typical flow
//
//  1. Build a [DataBuilder] with [New].
//  2. Register builder functions with [DataBuilder.AddBuilders].
//  3. Call [DataBuilder.Compile] with zero-valued instances of the structs
//     the caller will supply at runtime. Compile topologically sorts the
//     builders into stages, returning a [Plan].
//  4. Run the plan with [Plan.Run] (sequential) or [Plan.RunParallel]
//     (bounded worker pool). Both return a [Result].
//  5. Read typed outputs from the result with [Result.Get] or
//     [GetFromResult] from inside a builder.
//
// A compiled [Plan] is side-effect free and safe to reuse across goroutines.
// [Plan.Replace] can swap a builder for a compatible one without recompiling,
// as long as the replacement's inputs are a subset of the original's.
//
// # Parallelism
//
// [Plan.RunParallel] runs all builders in the same stage of the DAG
// concurrently, bounded by a caller-supplied worker count. A panic or error
// from any builder is surfaced back to the caller; subsequent stages do not
// start. Use [MaxPlanParallelism] to size the worker pool to the widest
// stage.
//
// # Performance
//
// Function-name (runtime.FuncForPC) and struct-name (reflect.Type)
// resolutions are cached in process-global sync.Maps. Keys are stable for
// the life of the program, so the caches never evict. Hot-path effects
// (benchstat, count=6):
//
//   - Result.Get: ~4x faster single-threaded, ~11x faster under parallel
//     load, zero allocations on hit.
//   - AddBuilders (warm cache): ~40% faster, ~60% fewer allocations.
//   - Per-resolution hits: ~10-15 ns/op, zero allocations.
//
// Benchmarks live in benchmarks_test.go; run `make bench` to measure on your
// hardware.
//
// # Visualization
//
// [BuildGraph] renders the compiled plan to a graphviz file in any format
// graphviz supports (png, svg, dot, ...). Graphviz must be installed on the
// system.
package databuilder
