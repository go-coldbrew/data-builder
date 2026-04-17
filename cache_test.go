package databuilder

// resetCachesForTest clears both resolution caches in place. It is safe only
// when no other goroutines are reading or writing the caches (i.e. from
// tests/benchmarks that are not running alongside live callers).
func resetCachesForTest() {
	structNameCache.Range(func(key, _ any) bool {
		structNameCache.Delete(key)
		return true
	})
	funcNameCache.Range(func(key, _ any) bool {
		funcNameCache.Delete(key)
		return true
	})
}
