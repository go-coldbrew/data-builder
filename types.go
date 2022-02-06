package databuilder

import "errors"

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
)

type DataBuilder interface {
	AddBuilders(fn ...interface{}) error
	Compile(initialData ...interface{}) (Plan, error)
}

type Plan interface {
	Run(...interface{}) (Data, error)
}

type Data map[string]interface{}
