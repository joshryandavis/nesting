package tart

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gitlab.com/gitlab-org/fleeting/nesting/hypervisor"
	"gitlab.com/gitlab-org/fleeting/nesting/hypervisor/internal/hvutil"
	"gitlab.com/gitlab-org/fleeting/nesting/hypervisor/tart/internal/control"
)

const (
	hvInitTimeout    = time.Minute
	vmAddressTimeout = 5 * time.Minute

	vmNamePrefix = "nesting-"
)

type Tart struct {
	mu  sync.Mutex
	vms map[string]func()
	cfg Config
}

type Config struct {
}

func New(config []byte) (*Tart, error) {
	hv := &Tart{
		vms: make(map[string]func()),
	}

	if len(config) > 0 {
		if err := json.Unmarshal(config, &hv.cfg); err != nil {
			return nil, fmt.Errorf("invalid config: %w", err)
		}
	}

	return hv, nil
}

func (hv *Tart) Init(ctx context.Context, config []byte) error {
	if len(config) > 0 {
		if err := json.Unmarshal(config, &hv.cfg); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, hvInitTimeout)
	defer cancel()

	return nil
}

func (hv *Tart) Shutdown(ctx context.Context) error {
	return nil
}

func (hv *Tart) Create(ctx context.Context, name string) (vm hypervisor.VirtualMachine, err error) {
	id, err := hvutil.UniqueID()
	if err != nil {
		return nil, fmt.Errorf("generating unique id: %w", err)
	}

	opts := control.CreateOptions{
		Id:      vmNamePrefix + id,
		Name:    name,
		Timeout: vmAddressTimeout,
	}

	var shutdown func()

	defer func() {
		if err == nil {
			return
		}

		if shutdown != nil {
			shutdown()
		}
		control.VirtualMachineDelete(context.Background(), opts.Id)
	}()

	shutdown, err = control.VirtualMachineCreate(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("starting vm: %w", err)
	}

	hv.mu.Lock()
	hv.vms[opts.Id] = shutdown
	hv.mu.Unlock()

	ipAddr, err := control.VirtualMachineAddress(ctx, opts.Id, vmAddressTimeout)
	if err != nil {
		return nil, err
	}

	return hypervisor.VirtualMachineInfo{
		Id:   opts.Id,
		Name: name,
		Addr: ipAddr,
	}, nil
}

func (hv *Tart) Delete(ctx context.Context, id string) error {
	hv.mu.Lock()
	if shutdown, ok := hv.vms[id]; ok {
		shutdown()
	}
	delete(hv.vms, id)
	hv.mu.Unlock()

	items, err := control.VirtualMachineList(ctx, id)
	if err != nil {
		return fmt.Errorf("fetching vm (%v) details: %w", id, err)
	}

	if len(items) == 0 {
		return fmt.Errorf("no vm (%v) found", id)
	}

	vm := items[0]
	if err := control.VirtualMachineDelete(ctx, vm); err != nil {
		return fmt.Errorf("stopping vm (%v): %w", id, err)
	}

	return nil
}

func (hv *Tart) List(ctx context.Context) ([]hypervisor.VirtualMachine, error) {
	items, err := control.VirtualMachineList(ctx, vmNamePrefix)
	if err != nil {
		return nil, fmt.Errorf("fetching list: %w", err)
	}

	vms := make([]hypervisor.VirtualMachine, 0, len(items))
	for _, item := range items {
		addr, err := control.VirtualMachineAddress(ctx, item, time.Second)
		if err != nil {
			return nil, fmt.Errorf("getting %q addr: %w", item, err)
		}

		vms = append(vms, hypervisor.VirtualMachineInfo{
			Id:   item,
			Addr: addr,
		})
	}

	return vms, nil
}
