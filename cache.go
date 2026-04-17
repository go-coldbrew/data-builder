package databuilder

import (
	"reflect"
	"runtime"
	"sync"
)

// Keys (reflect.Type identity, function PC) are stable for the lifetime of
// the process, so these caches never need eviction and are bounded by the
// number of distinct types and builder functions ever observed.
var (
	structNameCache sync.Map // reflect.Type -> string
	funcNameCache   sync.Map // uintptr      -> string
)

func cachedStructName(t reflect.Type) string {
	if v, ok := structNameCache.Load(t); ok {
		return v.(string)
	}
	name := t.PkgPath() + "." + t.Name()
	actual, _ := structNameCache.LoadOrStore(t, name)
	return actual.(string)
}

func resolveFuncName(pc uintptr) string {
	if v, ok := funcNameCache.Load(pc); ok {
		return v.(string)
	}
	name := runtime.FuncForPC(pc).Name()
	actual, _ := funcNameCache.LoadOrStore(pc, name)
	return actual.(string)
}

func resetCachesForTest() {
	structNameCache = sync.Map{}
	funcNameCache = sync.Map{}
}
