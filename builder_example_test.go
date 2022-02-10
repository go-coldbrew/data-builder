package databuilder

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

func ExampleDataBuilder() {
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

	//Output:
	// true
	// true
	// true
	// true
	// true
	// hello ankur!
	// welcome to singapore
}
