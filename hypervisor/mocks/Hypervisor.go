// Code generated by mockery v2.15.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	hypervisor "gitlab.com/gitlab-org/fleeting/nesting/hypervisor"
)

// Hypervisor is an autogenerated mock type for the Hypervisor type
type Hypervisor struct {
	mock.Mock
}

type Hypervisor_Expecter struct {
	mock *mock.Mock
}

func (_m *Hypervisor) EXPECT() *Hypervisor_Expecter {
	return &Hypervisor_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: ctx, name
func (_m *Hypervisor) Create(ctx context.Context, name string) (hypervisor.VirtualMachine, error) {
	ret := _m.Called(ctx, name)

	var r0 hypervisor.VirtualMachine
	if rf, ok := ret.Get(0).(func(context.Context, string) hypervisor.VirtualMachine); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(hypervisor.VirtualMachine)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Hypervisor_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type Hypervisor_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
func (_e *Hypervisor_Expecter) Create(ctx interface{}, name interface{}) *Hypervisor_Create_Call {
	return &Hypervisor_Create_Call{Call: _e.mock.On("Create", ctx, name)}
}

func (_c *Hypervisor_Create_Call) Run(run func(ctx context.Context, name string)) *Hypervisor_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Hypervisor_Create_Call) Return(_a0 hypervisor.VirtualMachine, _a1 error) *Hypervisor_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Delete provides a mock function with given fields: ctx, id
func (_m *Hypervisor) Delete(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Hypervisor_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type Hypervisor_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *Hypervisor_Expecter) Delete(ctx interface{}, id interface{}) *Hypervisor_Delete_Call {
	return &Hypervisor_Delete_Call{Call: _e.mock.On("Delete", ctx, id)}
}

func (_c *Hypervisor_Delete_Call) Run(run func(ctx context.Context, id string)) *Hypervisor_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Hypervisor_Delete_Call) Return(_a0 error) *Hypervisor_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

// Init provides a mock function with given fields: ctx, config
func (_m *Hypervisor) Init(ctx context.Context, config []byte) error {
	ret := _m.Called(ctx, config)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []byte) error); ok {
		r0 = rf(ctx, config)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Hypervisor_Init_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Init'
type Hypervisor_Init_Call struct {
	*mock.Call
}

// Init is a helper method to define mock.On call
//   - ctx context.Context
//   - config []byte
func (_e *Hypervisor_Expecter) Init(ctx interface{}, config interface{}) *Hypervisor_Init_Call {
	return &Hypervisor_Init_Call{Call: _e.mock.On("Init", ctx, config)}
}

func (_c *Hypervisor_Init_Call) Run(run func(ctx context.Context, config []byte)) *Hypervisor_Init_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]byte))
	})
	return _c
}

func (_c *Hypervisor_Init_Call) Return(_a0 error) *Hypervisor_Init_Call {
	_c.Call.Return(_a0)
	return _c
}

// List provides a mock function with given fields: ctx
func (_m *Hypervisor) List(ctx context.Context) ([]hypervisor.VirtualMachine, error) {
	ret := _m.Called(ctx)

	var r0 []hypervisor.VirtualMachine
	if rf, ok := ret.Get(0).(func(context.Context) []hypervisor.VirtualMachine); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]hypervisor.VirtualMachine)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Hypervisor_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type Hypervisor_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - ctx context.Context
func (_e *Hypervisor_Expecter) List(ctx interface{}) *Hypervisor_List_Call {
	return &Hypervisor_List_Call{Call: _e.mock.On("List", ctx)}
}

func (_c *Hypervisor_List_Call) Run(run func(ctx context.Context)) *Hypervisor_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Hypervisor_List_Call) Return(_a0 []hypervisor.VirtualMachine, _a1 error) *Hypervisor_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Shutdown provides a mock function with given fields: ctx
func (_m *Hypervisor) Shutdown(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Hypervisor_Shutdown_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Shutdown'
type Hypervisor_Shutdown_Call struct {
	*mock.Call
}

// Shutdown is a helper method to define mock.On call
//   - ctx context.Context
func (_e *Hypervisor_Expecter) Shutdown(ctx interface{}) *Hypervisor_Shutdown_Call {
	return &Hypervisor_Shutdown_Call{Call: _e.mock.On("Shutdown", ctx)}
}

func (_c *Hypervisor_Shutdown_Call) Run(run func(ctx context.Context)) *Hypervisor_Shutdown_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Hypervisor_Shutdown_Call) Return(_a0 error) *Hypervisor_Shutdown_Call {
	_c.Call.Return(_a0)
	return _c
}

type mockConstructorTestingTNewHypervisor interface {
	mock.TestingT
	Cleanup(func())
}

// NewHypervisor creates a new instance of Hypervisor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewHypervisor(t mockConstructorTestingTNewHypervisor) *Hypervisor {
	mock := &Hypervisor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
