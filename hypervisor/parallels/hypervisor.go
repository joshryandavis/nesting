package parallels

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"gitlab.com/ajwalker/nesting/hypervisor"
	"gitlab.com/ajwalker/nesting/hypervisor/internal/hvutil"
	"gitlab.com/ajwalker/nesting/hypervisor/parallels/internal/control"
)

const (
	hvInitTimeout    = time.Minute
	vmAddressTimeout = 5 * time.Minute

	vmNamePrefix      = "nesting-"
	networkNamePrefix = "isolation-"
)

type Parallels struct {
	mu       sync.Mutex
	networks map[string]bool
	cfg      Config
}

type Config struct {
	ImageDirectory   string `json:"image_directory"`
	WorkingDirectory string `json:"working_directory"`
	LicenseKey       string `json:"license_key"`
}

func New(config []byte) (*Parallels, error) {
	hv := &Parallels{}

	if len(config) > 0 {
		if err := json.Unmarshal(config, &hv.cfg); err != nil {
			return nil, fmt.Errorf("invalid config: %w", err)
		}
	}

	return hv, nil
}

func (hv *Parallels) Init(ctx context.Context, config []byte) error {
	if len(config) > 0 {
		if err := json.Unmarshal(config, &hv.cfg); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, hvInitTimeout)
	defer cancel()

	if err := control.InstallLicense(ctx, hv.cfg.LicenseKey); err != nil {
		return err
	}

	if err := hv.populateNetworks(ctx); err != nil {
		return fmt.Errorf("populating networks: %w", err)
	}

	return nil
}

func (hv *Parallels) Shutdown(ctx context.Context) error {
	return control.RemoveLicense(ctx)
}

func (hv *Parallels) Create(ctx context.Context, name string) (vm hypervisor.VirtualMachine, err error) {
	network, err := hv.getNetwork()
	if err != nil {
		return nil, err
	}

	var id, mac string
	if id, err = hvutil.UniqueID(); err != nil {
		return nil, fmt.Errorf("generating unique id: %w", err)
	}
	if mac, err = hvutil.GenerateMAC(); err != nil {
		return nil, fmt.Errorf("generating mac address: %w", err)
	}

	opts := control.CreateOptions{
		Id:         vmNamePrefix + id,
		ImagePath:  filepath.Join(hv.cfg.ImageDirectory, name+".pvm"),
		MAC:        mac,
		Network:    network,
		WorkingDir: hv.cfg.WorkingDirectory,
	}

	defer func() {
		if err != nil {
			control.VirtualMachineDelete(context.Background(), opts.Id)
			hv.putNetwork(network)
		}
	}()

	hv.mu.Lock()
	err = control.VirtualMachineCreate(ctx, opts)
	hv.mu.Unlock()

	if err != nil {
		return nil, fmt.Errorf("starting vm: %w", err)
	}

	ipAddr, err := getAddress(ctx, opts.MAC, vmAddressTimeout)
	if err != nil {
		return nil, err
	}

	return hypervisor.VirtualMachineInfo{
		Id:   opts.Id,
		Name: name,
		Addr: ipAddr,
	}, nil
}

func (hv *Parallels) Delete(ctx context.Context, id string) error {
	items, err := control.VirtualMachineList(ctx, id)
	if err != nil {
		return fmt.Errorf("fetching vm (%v) details: %w", id, err)
	}

	if len(items) == 0 {
		return fmt.Errorf("no vm (%v) found", id)
	}

	vm := items[0]
	if err := control.VirtualMachineDelete(ctx, vm.Name); err != nil {
		return fmt.Errorf("stopping vm (%v): %w", id, err)
	}

	// make network available
	hv.putNetwork(vm.Hardware.Net0.Iface)

	// remove dhcp lease
	removeLease(vm.Hardware.Net0.Mac)

	return nil
}

func (hv *Parallels) List(ctx context.Context) ([]hypervisor.VirtualMachine, error) {
	items, err := control.VirtualMachineList(ctx, vmNamePrefix)
	if err != nil {
		return nil, fmt.Errorf("fetching list: %w", err)
	}

	vms := make([]hypervisor.VirtualMachine, 0, len(items))
	for _, item := range items {
		addr, err := getAddress(ctx, item.Hardware.Net0.Mac, 0)
		if err != nil {
			return nil, fmt.Errorf("getting vm addr: %w", err)
		}

		vms = append(vms, hypervisor.VirtualMachineInfo{
			Id:   item.Name,
			Name: item.Description,
			Addr: addr,
		})
	}

	return vms, nil
}
