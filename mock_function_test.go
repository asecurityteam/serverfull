// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/asecurityteam/serverfull (interfaces: Function)

// Package serverfull is a generated GoMock package.
package serverfull

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockFunction is a mock of Function interface
type MockFunction struct {
	ctrl     *gomock.Controller
	recorder *MockFunctionMockRecorder
}

// MockFunctionMockRecorder is the mock recorder for MockFunction
type MockFunctionMockRecorder struct {
	mock *MockFunction
}

// NewMockFunction creates a new mock instance
func NewMockFunction(ctrl *gomock.Controller) *MockFunction {
	mock := &MockFunction{ctrl: ctrl}
	mock.recorder = &MockFunctionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFunction) EXPECT() *MockFunctionMockRecorder {
	return m.recorder
}

// Errors mocks base method
func (m *MockFunction) Errors() []error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Errors")
	ret0, _ := ret[0].([]error)
	return ret0
}

// Errors indicates an expected call of Errors
func (mr *MockFunctionMockRecorder) Errors() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Errors", reflect.TypeOf((*MockFunction)(nil).Errors))
}

// Invoke mocks base method
func (m *MockFunction) Invoke(arg0 context.Context, arg1 []byte) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Invoke", arg0, arg1)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Invoke indicates an expected call of Invoke
func (mr *MockFunctionMockRecorder) Invoke(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Invoke", reflect.TypeOf((*MockFunction)(nil).Invoke), arg0, arg1)
}

// Source mocks base method
func (m *MockFunction) Source() interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Source")
	ret0, _ := ret[0].(interface{})
	return ret0
}

// Source indicates an expected call of Source
func (mr *MockFunctionMockRecorder) Source() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Source", reflect.TypeOf((*MockFunction)(nil).Source))
}
