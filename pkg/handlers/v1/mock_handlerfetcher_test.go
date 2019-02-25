// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/asecurityteam/serverfull/pkg/domain (interfaces: HandlerFetcher)

// Package v1 is a generated GoMock package.
package v1

import (
	context "context"
	reflect "reflect"

	lambda "github.com/aws/aws-lambda-go/lambda"
	gomock "github.com/golang/mock/gomock"
)

// MockHandlerFetcher is a mock of HandlerFetcher interface
type MockHandlerFetcher struct {
	ctrl     *gomock.Controller
	recorder *MockHandlerFetcherMockRecorder
}

// MockHandlerFetcherMockRecorder is the mock recorder for MockHandlerFetcher
type MockHandlerFetcherMockRecorder struct {
	mock *MockHandlerFetcher
}

// NewMockHandlerFetcher creates a new mock instance
func NewMockHandlerFetcher(ctrl *gomock.Controller) *MockHandlerFetcher {
	mock := &MockHandlerFetcher{ctrl: ctrl}
	mock.recorder = &MockHandlerFetcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHandlerFetcher) EXPECT() *MockHandlerFetcherMockRecorder {
	return m.recorder
}

// FetchHandler mocks base method
func (m *MockHandlerFetcher) FetchHandler(arg0 context.Context, arg1 string) (lambda.Handler, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchHandler", arg0, arg1)
	ret0, _ := ret[0].(lambda.Handler)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchHandler indicates an expected call of FetchHandler
func (mr *MockHandlerFetcherMockRecorder) FetchHandler(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchHandler", reflect.TypeOf((*MockHandlerFetcher)(nil).FetchHandler), arg0, arg1)
}
