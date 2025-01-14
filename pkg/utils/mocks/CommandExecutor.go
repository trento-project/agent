// Code generated by mockery v2.32.3. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// CommandExecutor is an autogenerated mock type for the CommandExecutor type
type CommandExecutor struct {
	mock.Mock
}

// Exec provides a mock function with given fields: name, arg
func (_m *CommandExecutor) Exec(name string, arg ...string) ([]byte, error) {
	_va := make([]interface{}, len(arg))
	for _i := range arg {
		_va[_i] = arg[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, name)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(string, ...string) ([]byte, error)); ok {
		return rf(name, arg...)
	}
	if rf, ok := ret.Get(0).(func(string, ...string) []byte); ok {
		r0 = rf(name, arg...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(string, ...string) error); ok {
		r1 = rf(name, arg...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ExecContext provides a mock function with given fields: ctx, name, arg
func (_m *CommandExecutor) ExecContext(ctx context.Context, name string, arg ...string) ([]byte, error) {
	_va := make([]interface{}, len(arg))
	for _i := range arg {
		_va[_i] = arg[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, name)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...string) ([]byte, error)); ok {
		return rf(ctx, name, arg...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, ...string) []byte); ok {
		r0 = rf(ctx, name, arg...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, ...string) error); ok {
		r1 = rf(ctx, name, arg...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewCommandExecutor creates a new instance of CommandExecutor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCommandExecutor(t interface {
	mock.TestingT
	Cleanup(func())
}) *CommandExecutor {
	mock := &CommandExecutor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
