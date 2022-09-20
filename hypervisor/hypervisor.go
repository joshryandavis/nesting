package hypervisor

import (
	"context"
)

type Hypervisor interface {
	Init(ctx context.Context, config []byte) error
	Shutdown(ctx context.Context) error

	Create(ctx context.Context, name string) (VirtualMachine, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]VirtualMachine, error)
}

type VirtualMachine interface {
	GetId() string
	GetName() string
	GetAddr() string
}

type VirtualMachineInfo struct {
	Id   string
	Name string
	Addr string
}

func (vmi VirtualMachineInfo) GetId() string {
	return vmi.Id
}

func (vmi VirtualMachineInfo) GetName() string {
	return vmi.Name
}

func (vmi VirtualMachineInfo) GetAddr() string {
	return vmi.Addr
}
