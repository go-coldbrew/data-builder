package databuilder

import (
	"context"
	"errors"
)

var (
	// ErrInvalidBuilder is returned when the builder is not valid
	ErrInvalidBuilder = errors.New("The provided builder is invalid")
	// ErrInvalidBuilderKind is returned when the builder is not a function
	ErrInvalidBuilderKind = errors.New("invalid builder, should only be a function")
	// ErrInvalidBuilderNumInput is returned when the builder does not have 1 input
	ErrInvalidBuilderNumOutput = errors.New("invalid builder, should always return two values")
	// ErrInvalidBuilderFirstOutput is returned when the builder does not return a struct as first output
	ErrInvalidBuilderFirstOutput = errors.New("invalid builder, first return type should be a struct")
	// ErrInvalidBuilderSecondOutput is returned when the builder does not return an error as second output
	ErrInvalidBuilderSecondOutput = errors.New("invalid builder, second return type should be error")
	// ErrInvalidBuilderMissingContext is returned when the builder does not have a context as first input
	ErrInvalidBuilderMissingContext = errors.New("invalid builder, missing context")
	// ErrInvalidBuilderInput is returned when the builder does not have a struct as input
	ErrInvalidBuilderInput = errors.New("invalid builder, input should be a struct")
	// ErrInvalidBuilderOutput is returned when the builder does not have a struct as output
	ErrMultipleBuilderSameOutput = errors.New("invalid, multiple builders CAN NOT produce the same output")
	// ErrSameInputAsOutput is returned when the builder has the same input and output
	ErrSameInputAsOutput = errors.New("invalid builder, input and output should NOT be same")
	// ErrCouldNotResolveDependency is returned when the builder can not be resolved
	ErrCouldNotResolveDependency = errors.New("dependency can not be resolved")
	// ErrMultipleInitialData is returned when the initial data is provided twice
	ErrMultipleInitialData = errors.New("initial data provided twice")
	// ErrInitialDataMissing is returned when the initial data is not provided
	ErrInitialDataMissing = errors.New("need complile time defined initial data to run")
)

// DataBuilder is the interface for DataBuilder
type DataBuilder interface {
	// AddBuilders adds the builders to the DataBuilder. The builders are added to the DataBuilder
	AddBuilders(fn ...interface{}) error
	// Compile compiles the builders and returns a plan that can be used to run the builders
	// The initial data is used to resolve the dependencies of the builders. The initial data should be a struct that contains the fields that are used as input for the builders when this Plan is executed.
	Compile(initialData ...interface{}) (Plan, error)
}

// Plan is the interface that wraps execution of Plans created by DataBuilder.Compile method.
type Plan interface {
	// Replace replaces the builder function used in compile with a different function. The builder function should be the same as the one used in AddBuilders
	Replace(ctx context.Context, from, to interface{}) error
	// Run runs the builders in the plan. The initial data is used to resolve the dependencies of the builders. The initial data should be a struct that contains the fields that are used as input for the builders when this Plan is executed.
	Run(ctx context.Context, initValues ...interface{}) (Result, error)
	// RunParallel runs the builders in the plan in parallel. The initial data is used to resolve the dependencies of the builders. The initial data should be a struct that contains the fields that are used as input for the builders when this Plan is executed.
	RunParallel(ctx context.Context, count uint, initValues ...interface{}) (Result, error)
}

// Result is the result of the Plan.Run method
type Result map[string]interface{}
