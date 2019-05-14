package serverfull

import (
	"github.com/aws/aws-lambda-go/lambda"
)

// LambdaFunction is a small wrapper around the lambda.Handler
// that preserves the original signature of the function for later
// retrieval.
type LambdaFunction struct {
	lambda.Handler
	source interface{}
	errors []error
}

// Source returns the original function signature.
func (f *LambdaFunction) Source() interface{} {
	return f.source
}

// Errors returns a list of errors the Lambda might return. This is
// only populated if the function was constructed using the
// NewFunctionWithErrors constructor.
func (f *LambdaFunction) Errors() []error {
	return f.errors
}

// NewFunctionWithErrors allows for documenting the various error types that
// can be returned by the function. This may be used when running in mock + http
// build modes to trigger exceptions.
func NewFunctionWithErrors(v interface{}, errors ...error) Function {
	return &LambdaFunction{
		Handler: lambda.NewHandler(v),
		source:  v,
		errors:  errors,
	}
}

// NewFunction is a replacement for lambda.NewHandler that returns
// a Function.
func NewFunction(v interface{}) Function {
	return &LambdaFunction{
		Handler: lambda.NewHandler(v),
		source:  v,
	}
}
