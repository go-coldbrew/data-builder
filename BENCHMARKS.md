# Benchmarks: Function / Type Name Resolution Caching

`data-builder` resolves two "locations" on every request:

1. **PC → function name** via `runtime.FuncForPC(pc).Name()`
2. **`reflect.Type` → qualified struct name** via `t.PkgPath() + "." + t.Name()`

Both are now cached in process-global `sync.Map`s (see `cache.go`). Keys
(reflect.Type identity, function PC) are stable for the life of the program,
so the caches never need eviction.

## Reproducing

```sh
go install golang.org/x/perf/cmd/benchstat@latest

# "before": with cache.go rewritten to a pass-through (no caching)
go test -run=^$ -bench=. -benchmem -count=6 ./... | tee before.txt

# "after": with cache.go in its cached form
go test -run=^$ -bench=. -benchmem -count=6 ./... | tee after.txt

benchstat before.txt after.txt
```

The benchmark suite lives in `benchmarks_test.go`. `make bench` runs it with
`-count=1`; use the commands above for statistically stable comparisons.

## Environment

- `go version go1.25.8 linux/amd64`
- CPU: INTEL(R) XEON(R) PLATINUM 8581C @ 2.10GHz (16 logical cores)
- Kernel: Linux 4.4.0
- `benchstat` with `-count=6`

## Results (benchstat)

### Time per op

| Benchmark                    |   Before |    After |          Δ |
| ---------------------------- | -------: | -------: | ---------: |
| `GetStructName_Uncached`     |  81.28ns |  84.98ns |         ~  |
| `CachedStructName_Hit`       |  83.79ns |  11.23ns | **-86.6%** |
| `CachedStructName_ColdMix`   |  86.16ns |  11.46ns | **-86.7%** |
| `FuncForPC_Uncached`         |  32.71ns |  33.38ns |         ~  |
| `ResolveFuncName_Hit`        |  32.60ns |  12.07ns | **-63.0%** |
| `ResolveFuncName_ColdMix`    |  30.90ns |  12.13ns | **-60.7%** |
| `AddBuilders`                |  3.950µs |  2.300µs | **-41.8%** |
| `AddBuilders_ColdCache`      |  8.089µs | 10.357µs |    +28.1%  |
| `Compile`                    |  6.920µs |  7.006µs |         ~  |
| `RunParallel_Workers1`       |  15.64µs |  15.44µs |         ~  |
| `RunParallel_Workers4`       |  20.71µs |  20.71µs |         ~  |
| `RunParallel_Workers8`       |  23.91µs |  23.59µs |         ~  |
| `ResultGet`                  | 103.70ns |  25.54ns | **-75.4%** |
| `ResultGet_Parallel`         |  16.80ns |   1.54ns | **-90.8%** |
| **geomean**                  |   498ns  |   244ns  | **-51.0%** |

### Allocations

| Benchmark               | Before   | After   | Δ B/op     | Δ allocs/op |
| ----------------------- | -------: | ------: | ---------: | ----------: |
| `CachedStructName_Hit`  |      48B |      0B |   **-100%** |   **-100%** |
| `CachedStructName_ColdMix` |   51B |      0B |   **-100%** |   **-100%** |
| `AddBuilders`           |   1872B |    928B |    -50.4%  |    -59.4%   |
| `Compile`               |   4328B |   4266B |     -1.4%  |     -2.2%   |
| `RunParallel_Workers1`  |   4945B |   4695B |     -5.1%  |     -6.3%   |
| `RunParallel_Workers4`  |   5036B |   4786B |     -5.0%  |     -6.1%   |
| `RunParallel_Workers8`  |   5161B |   4911B |     -4.8%  |     -5.8%   |
| `ResultGet`             |     48B |      0B |   **-100%** |   **-100%** |
| `ResultGet_Parallel`    |     48B |      0B |   **-100%** |   **-100%** |

Statistical significance: all reported deltas have `p=0.002` with n=6; entries
marked `~` are not statistically distinguishable from the baseline.

## Interpretation

**Where caching helps most**

- `Result.Get` and the hot path inside `doWorkAndGetResult` / `RunParallel`
  init loops used to allocate a fresh `string` for every type lookup
  (`t.PkgPath() + "." + t.Name()`). Interning the result via `sync.Map`
  eliminates that allocation entirely: `ResultGet` drops 78ns and one
  allocation; under parallel load (`ResultGet_Parallel`) it goes from
  16.8ns to 1.5ns — an **11× speedup** because `sync.Map`'s read-only
  fast-path is lock-free and scales linearly across cores.
- `AddBuilders` gets a steady-state 42% latency win and 59% fewer
  allocations because each builder registration re-resolves the same input
  and output type names several times via `IsValidBuilder` and `getBuilder`.
- `FuncForPC` caching is a smaller absolute win (20ns / call) than struct
  name caching, but it's on the same hot path for `getBuilder` and
  `plan.Replace`, so it still helps `AddBuilders` directly.

**Where caching does not help (and that's fine)**

- `Compile` and `RunParallel` end-to-end are dominated by
  `resolveDependencies`, goroutine scheduling, and `reflect.Value.Call`.
  Name resolution is <5% of those timings, so benchstat reports "no
  significant change" — but the memory column still shows a real reduction
  (~5% bytes/allocs per run) because those allocations were shifted off
  the hot path.
- `_Uncached` baselines for both resolvers come in identical before and
  after (as expected — they call the un-cached code directly).

**The `AddBuilders_ColdCache` regression**

This synthetic benchmark resets both `sync.Map`s to empty at the start of
every iteration, so every call is a miss. `sync.Map` is slower than a
direct computation in the pure-miss case because it pays for an atomic
`Load` + an `LoadOrStore` on top of the original work. In production the
cache warms up once and then serves hits forever, so this scenario isn't
observable in practice — it's included only to pin the worst-case cost.

## Caveats

- `sync.Map` has higher per-op overhead than a plain `map` when the working
  set is tiny **and** purely single-threaded. The `_ColdMix` benchmarks
  are intentionally small (5 types / 4 PCs) to stress this path; they
  still show ~85-87% wins, because `PkgPath()+Name()` and `FuncForPC`
  dominate the miss cost.
- Absolute numbers depend on CPU, OS scheduler, and the number of distinct
  types/builders the program touches. Don't generalize — re-measure in
  the target deployment if it matters.
- Benchmarks should be run with the machine idle; pin `GOMAXPROCS` if you
  want tighter variance across runs.
