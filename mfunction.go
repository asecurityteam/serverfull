package serverfull

import (
	"context"
	"reflect"
)

// MockingFetcher sources original functions from another Fetcher
// and mocks out the results.
type MockingFetcher struct {
	Fetcher Fetcher
}

// Fetch calls the underlying Fetcher and mocks the results.
func (f *MockingFetcher) Fetch(ctx context.Context, name string) (Function, error) {
	r, err := f.Fetcher.Fetch(ctx, name)
	if err != nil {
		return nil, err
	}
	return mockFunction(r), nil
}

func mockFunction(f Function) Function {
	// Because the function previously passed validation by
	// the official lambda SDK then we will assume a few characteristics
	// of the function. Notably, we assume that any non-zero return
	// values from the function means that the last return value is an error.
	t := reflect.TypeOf(f.Source())
	out := t.NumOut()
	returnsError := out == 1 || out == 2
	var returnType reflect.Type
	if out == 2 {
		returnType = t.Out(0)
	}
	mockFn := newMockFn(returnType, returnsError)
	newFn := reflect.MakeFunc(t, mockFn)
	return NewFunctionWithErrors(
		newFn.Interface(),
		f.Errors()...,
	)
}

func newMockFn(returnType reflect.Type, returnsError bool) func(args []reflect.Value) []reflect.Value {
	return func(_ []reflect.Value) []reflect.Value {
		res := make([]reflect.Value, 0)
		if returnType != nil {
			res = append(res, reflect.Indirect(reflect.New(returnType)))
		}
		if returnsError {
			res = append(res, reflect.Zero(reflect.TypeOf((*error)(nil)).Elem()))
		}
		return res
	}
}
