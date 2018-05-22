package mocks

import (
	gomock "github.com/golang/mock/gomock"

	"reflect"
)

// MockContract is a mock of Contract interface
type MockContract struct {
	ctrl     *gomock.Controller
	recorder *MockContractMockRecorder
}

// MockContractMockRecorder is the mock recorder for MockContract
type MockContractMockRecorder struct {
	mock *MockContract
}

// NewMockContract creates a new mock instance
func NewMockContract(ctrl *gomock.Controller) *MockContract {
	mock := &MockContract{ctrl: ctrl}
	mock.recorder = &MockContractMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockContract) EXPECT() *MockContractMockRecorder {
	return m.recorder
}

// API mocks base method
func (m *MockContract) API(arg0 string) (vm.Method, error) {
	ret := m.ctrl.Call(m, "API", arg0)
	ret0, _ := ret[0].(vm.Method)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// API indicates an expected call of API
func (mr *MockContractMockRecorder) API(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "API", reflect.TypeOf((*MockContract)(nil).API), arg0)
}

// AddSigner mocks base method
func (m *MockContract) AddSigner(arg0 []byte) {
	m.ctrl.Call(m, "AddSigner", arg0)
}

// AddSigner indicates an expected call of AddSigner
func (mr *MockContractMockRecorder) AddSigner(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddSigner", reflect.TypeOf((*MockContract)(nil).AddSigner), arg0)
}

// Code mocks base method
func (m *MockContract) Code() string {
	ret := m.ctrl.Call(m, "Code")
	ret0, _ := ret[0].(string)
	return ret0
}