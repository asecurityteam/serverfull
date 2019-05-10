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
}

// Source returns the original function signature.
func (f *LambdaFunction) Source() interface{} {
	return f.source
}

// NewFunction is a replacement for lambda.NewHandler that returns
// a Function.
func NewFunction(v interface{}) Function {
	return &LambdaFunction{
		Handler: lambda.NewHandler(v),
		source:  v,
	}
}
