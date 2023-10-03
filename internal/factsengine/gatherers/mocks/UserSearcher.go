// Code generated by mockery v2.34.2. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// UserSearcher is an autogenerated mock type for the UserSearcher type
type UserSearcher struct {
	mock.Mock
}

// GetUsernameByID provides a mock function with given fields: userID
func (_m *UserSearcher) GetUsernameByID(userID string) (string, error) {
	ret := _m.Called(userID)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(userID)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(userID)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewUserSearcher creates a new instance of UserSearcher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewUserSearcher(t interface {
	mock.TestingT
	Cleanup(func())
}) *UserSearcher {
	mock := &UserSearcher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
