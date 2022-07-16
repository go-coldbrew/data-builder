package databuilder

import "context"

const (
	pKey = "github.com/coldebrew-go/data-builder.Result"
)

// AddResultToCtx adds the given result object to context
//
// this function should ideally only be used in your tests and/or for debugging
// modification made to Result obj will NOT persist
func AddResultToCtx(ctx context.Context, r Result) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, pKey, r)
}

// GetResultFromCtx gives access to result object at this point in execution
//
// this function should ideally only be used in your tests and/or for debugging
// modification made to Result obj may or may not persist
func GetResultFromCtx(ctx context.Context) Result {
	v := ctx.Value(pKey)
	if v == nil {
		return nil
	}
	if r, ok := v.(Result); ok {
		return r
	}
	return nil
}

// GetFromResult allows builders to access data built by other builders
//
// this function enables optional access to data, your code should not rely on
// values being present, if you have explicit dependency please add them to your
// function parameters
func GetFromResult(ctx context.Context, obj interface{}) interface{} {
	r := GetResultFromCtx(ctx)
	if r == nil {
		return nil
	}
	return r.Get(obj)
}
