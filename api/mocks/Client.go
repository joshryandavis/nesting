// Code generated by mockery v2.15.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	hypervisor "gitlab.com/gitlab-org/fleeting/nesting/hypervisor"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

type Client_Expecter struct {
	mock *mock.Mock
}

func (_m *Client) EXPECT() *Client_Expecter {
	return &Client_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with given fields:
func (_m *Client) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type Client_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *Client_Expecter) Close() *Client_Close_Call {
	return &Client_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *Client_Close_Call) Run(run func()) *Client_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Client_Close_Call) Return(_a0 error) *Client_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

// Create provides a mock function with given fields: ctx, name, slot
func (_m *Client) Create(ctx context.Context, name string, slot *int32) (hypervisor.VirtualMachine, *string, error) {
	ret := _m.Called(ctx, name, slot)

	var r0 hypervisor.VirtualMachine
	if rf, ok := ret.Get(0).(func(context.Context, string, *int32) hypervisor.VirtualMachine); ok {
		r0 = rf(ctx, name, slot)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(hypervisor.VirtualMachine)
		}
	}

	var r1 *string
	if rf, ok := ret.Get(1).(func(context.Context, string, *int32) *string); ok {
		r1 = rf(ctx, name, slot)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*string)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, string, *int32) error); ok {
		r2 = rf(ctx, name, slot)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Client_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type Client_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
//   - slot *int32
func (_e *Client_Expecter) Create(ctx interface{}, name interface{}, slot interface{}) *Client_Create_Call {
	return &Client_Create_Call{Call: _e.mock.On("Create", ctx, name, slot)}
}

func (_c *Client_Create_Call) Run(run func(ctx context.Context, name string, slot *int32)) *Client_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*int32))
	})
	return _c
}

func (_c *Client_Create_Call) Return(vm hypervisor.VirtualMachine, stompedVmId *string, err error) *Client_Create_Call {
	_c.Call.Return(vm, stompedVmId, err)
	return _c
}

// Delete provides a mock function with given fields: ctx, id
func (_m *Client) Delete(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type Client_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *Client_Expecter) Delete(ctx interface{}, id interface{}) *Client_Delete_Call {
	return &Client_Delete_Call{Call: _e.mock.On("Delete", ctx, id)}
}

func (_c *Client_Delete_Call) Run(run func(ctx context.Context, id string)) *Client_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Client_Delete_Call) Return(_a0 error) *Client_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

// Init provides a mock function with given fields: ctx, config
func (_m *Client) Init(ctx context.Context, config []byte) error {
	ret := _m.Called(ctx, config)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []byte) error); ok {
		r0 = rf(ctx, config)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_Init_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Init'
type Client_Init_Call struct {
	*mock.Call
}

// Init is a helper method to define mock.On call
//   - ctx context.Context
//   - config []byte
func (_e *Client_Expecter) Init(ctx interface{}, config interface{}) *Client_Init_Call {
	return &Client_Init_Call{Call: _e.mock.On("Init", ctx, config)}
}

func (_c *Client_Init_Call) Run(run func(ctx context.Context, config []byte)) *Client_Init_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]byte))
	})
	return _c
}

func (_c *Client_Init_Call) Return(_a0 error) *Client_Init_Call {
	_c.Call.Return(_a0)
	return _c
}

// List provides a mock function with given fields: ctx
func (_m *Client) List(ctx context.Context) ([]hypervisor.VirtualMachine, error) {
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

// Client_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type Client_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - ctx context.Context
func (_e *Client_Expecter) List(ctx interface{}) *Client_List_Call {
	return &Client_List_Call{Call: _e.mock.On("List", ctx)}
}

func (_c *Client_List_Call) Run(run func(ctx context.Context)) *Client_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Client_List_Call) Return(_a0 []hypervisor.VirtualMachine, _a1 error) *Client_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Shutdown provides a mock function with given fields: ctx
func (_m *Client) Shutdown(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_Shutdown_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Shutdown'
type Client_Shutdown_Call struct {
	*mock.Call
}

// Shutdown is a helper method to define mock.On call
//   - ctx context.Context
func (_e *Client_Expecter) Shutdown(ctx interface{}) *Client_Shutdown_Call {
	return &Client_Shutdown_Call{Call: _e.mock.On("Shutdown", ctx)}
}

func (_c *Client_Shutdown_Call) Run(run func(ctx context.Context)) *Client_Shutdown_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Client_Shutdown_Call) Return(_a0 error) *Client_Shutdown_Call {
	_c.Call.Return(_a0)
	return _c
}

type mockConstructorTestingTNewClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewClient creates a new instance of Client. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewClient(t mockConstructorTestingTNewClient) *Client {
	mock := &Client{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
