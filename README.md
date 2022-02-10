<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# databuilder

```go
import "github.com/go-coldbrew/data-builder"
```

## Index

- [Variables](<#variables>)
- [func AddResultToCtx(ctx context.Context, r Result) context.Context](<#func-addresulttoctx>)
- [func BuildGraph(executionPlan Plan, format, file string) error](<#func-buildgraph>)
- [func GetFromResult(ctx context.Context, obj interface{}) interface{}](<#func-getfromresult>)
- [func IsValidBuilder(builder interface{}) error](<#func-isvalidbuilder>)
- [type DataBuilder](<#type-databuilder>)
  - [func New() DataBuilder](<#func-new>)
- [type Plan](<#type-plan>)
- [type Result](<#type-result>)
  - [func GetResultFromCtx(ctx context.Context) Result](<#func-getresultfromctx>)
  - [func (r Result) Get(obj interface{}) interface{}](<#func-result-get>)


## Variables

```go
var (
    ErrInvalidBuilder               = errors.New("The provided builder is invalid")
    ErrInvalidBuilderKind           = errors.New("invalid builder, should only be a function")
    ErrInvalidBuilderNumOutput      = errors.New("invalid builder, should always return two values")
    ErrInvalidBuilderFirstOutput    = errors.New("invalid builder, first return type should be a struct")
    ErrInvalidBuilderSecondOutput   = errors.New("invalid builder, second return type should be error")
    ErrInvalidBuilderMissingContext = errors.New("invalid builder, missing context")
    ErrInvalidBuilderInput          = errors.New("invalid builder, input should be a struct")
    ErrMultipleBuilderSameOutput    = errors.New("invalid, multiple builders CAN NOT produce the same output")
    ErrSameInputAsOutput            = errors.New("invalid builder, input and output should NOT be same")
    ErrCouldNotResolveDependency    = errors.New("dependency can not be resolved")
    ErrMultipleInitialData          = errors.New("initial data provided twice")
    ErrInitialDataMissing           = errors.New("need complile time defined initial data to run")
)
```

## func AddResultToCtx

```go
func AddResultToCtx(ctx context.Context, r Result) context.Context
```

### AddResultToCtx adds the given result object to context

this function should ideally only be used in your tests and/or for debugging modification made to Result obj will NOT persist

## func BuildGraph

```go
func BuildGraph(executionPlan Plan, format, file string) error
```

BuildGraph helps understand the execution plan\, it renders the plan in the given format please note we depend on graphviz\, please ensure you have graphviz installed

## func GetFromResult

```go
func GetFromResult(ctx context.Context, obj interface{}) interface{}
```

### GetFromResult allows builders to access data built by other builders

this function enables optional access to data\, your code should not rely on values being present\, if you have explicit dependency please add them to your function parameters

## func IsValidBuilder

```go
func IsValidBuilder(builder interface{}) error
```

IsValidBuilder checks if the given function is valid or not

## type DataBuilder

```go
type DataBuilder interface {
    AddBuilders(fn ...interface{}) error
    Compile(initialData ...interface{}) (Plan, error)
}
```

<details><summary>Example</summary>
<p>

```go
package main

import (
	"context"
	"fmt"
	"strings"
)

// lets say we have some data being produced by a set of functions
// but we need to define how their interaction should be and how their dependency
// should be resolved

type AppRequest struct {
	FirstName string
	CityName  string
	UpperCase bool
	LowerCase bool
}

type AppResponse struct {
	Msg string
}

type NameMsg struct {
	Msg string
}

type CityMsg struct {
	Msg string
}

type CaseMsg struct {
	Msg string
}

// Lets try to build a sample builder with some dependency
// Assuming we have an App that acts on the request
// processes it in multiple steps and returns a Response
// we can think of this process as a series of functions

// NameMsgBuilder builds name salutation from our AppRequest
func NameMsgBuilder(_ context.Context, req AppRequest) (NameMsg, error) {
	return NameMsg{
		Msg: fmt.Sprintf("Hello %s!", req.FirstName),
	}, nil
}

//CityMsgBuilder builds city welcome msg from our AppRequest
func CityMsgBuilder(_ context.Context, req AppRequest) (CityMsg, error) {
	return CityMsg{
		Msg: fmt.Sprintf("Welcome to %s", req.CityName),
	}, nil
}

// CaseMsgBuilder handles the case transformation of the message
func CaseMsgBuilder(_ context.Context, name NameMsg, city CityMsg, req AppRequest) (CaseMsg, error) {
	msg := fmt.Sprintf("%s\n%s", name.Msg, city.Msg)
	if req.UpperCase {
		msg = strings.ToUpper(msg)
	} else if req.LowerCase {
		msg = strings.ToLower(msg)
	}
	return CaseMsg{
		Msg: msg,
	}, nil
}

//ResponseBuilder builds Application response from CaseMsg
func ResponseBuilder(_ context.Context, m CaseMsg) (AppResponse, error) {
	return AppResponse{
		Msg: m.Msg,
	}, nil
}

func main() {
	// First we build an object of the builder interface
	b := New()

	// Then we add all the builders
	// its okay to call `AddBuilders` multiple times
	err := b.AddBuilders(
		NameMsgBuilder,
		CityMsgBuilder,
		CaseMsgBuilder,
	)
	fmt.Println(err == nil)

	// lets ass all builders
	err = b.AddBuilders(ResponseBuilder)
	fmt.Println(err == nil)

	// next we we compile this into a plan
	// the compilation ensures we have a resolved dependency graph
	_, err = b.Compile()
	fmt.Println(err != nil)

	// Why did we get the error ?
	// if we look at our dependency graph, there is no builder that produces AppRequest
	// in order of dependency resolution to work we need to tell
	// the Compile method that we will provide it some initial Data

	// we can do that by passing empty structs
	// compiler just needs the type, values will come in later
	ep, err := b.Compile(AppRequest{})
	fmt.Println(err == nil)

	// once the Compilation has finished, we get an execution plan
	// the execution plan once created can be cached and is side effect free
	// It can be executed across multiple go routines
	// lets run the Plan, remember to pass in the initial value
	result, err := ep.Run(
		context.Background(), // context is passed on the builders
		AppRequest{
			FirstName: "Ankur",
			CityName:  "Singapore",
			LowerCase: true,
		},
	)
	fmt.Println(err == nil)

	// once the execution is done, we can read all the values from the result
	resp := AppResponse{}
	resp = result.Get(resp).(AppResponse)
	fmt.Println(resp.Msg)

}
```

#### Output

```
true
true
true
true
true
hello ankur!
welcome to singapore
```

</p>
</details>

### func New

```go
func New() DataBuilder
```

New Creates a new DataBuilder

## type Plan

```go
type Plan interface {
    Replace(ctx context.Context, from, to interface{}) error
    Run(ctx context.Context, initValues ...interface{}) (Result, error)
}
```

<details><summary>Example</summary>
<p>

```go
{
	b := New()
	err := b.AddBuilders(DBTestFunc, DBTestFunc4)
	fmt.Println(err == nil)
	ep, err := b.Compile(TestStruct1{})
	fmt.Println(err == nil)

	_, err = ep.Run(context.Background(), TestStruct1{})
	fmt.Println(err == nil)

	err = ep.Replace(context.Background(), DBTestFunc, DBTestFunc5)
	fmt.Println(err == nil)
	_, err = ep.Run(context.Background(), TestStruct1{})
	fmt.Println(err == nil)

}
```

#### Output

```
true
true
CALLED DBTestFunc
CALLED DBTestFunc4
true
true
CALLED DBTestFunc5
CALLED DBTestFunc4
true
```

</p>
</details>

## type Result

```go
type Result map[string]interface{}
```

### func GetResultFromCtx

```go
func GetResultFromCtx(ctx context.Context) Result
```

#### GetResultFromCtx gives access to result object at this point in execution

this function should ideally only be used in your tests and/or for debugging modification made to Result obj may or may not persist

### func \(Result\) Get

```go
func (r Result) Get(obj interface{}) interface{}
```



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
