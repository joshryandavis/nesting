//go:build darwin && arm64

package virtualizationframework

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitlab.com/gitlab-org/fleeting/nesting/hypervisor"
	"golang.org/x/sync/errgroup"

	"github.com/Code-Hex/vz/v3"
)

type VirtualizationFramework struct {
	mu  sync.Mutex
	vms map[string]virtualMachine
	cfg Config
}

type virtualMachine struct {
	id   string
	addr string
	name string

	vm       *vz.VirtualMachine
	shutdown func() error
}

type Config struct {
	ImageDirectory   string `json:"image_directory"`
	WorkingDirectory string `json:"working_directory"`
}

var errVirtualMachineStopped = errors.New("virtual machine stopped")

// VirtualMachineConfig is an indivual VM's configuration, this is modelled after
// the config Tart uses.
type VirtualMachineConfig struct {
	id string

	Version       int    `json:"version"`
	MemorySize    uint64 `json:"memorySize"`
	Arch          string `json:"arch"`
	OS            string `json:"os"`
	HardwareModel []byte `json:"hardwareModel"`
	CPUCount      uint   `json:"cpuCount"`
	Display       struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"display"`
	ECID       []byte `json:"ecid"`
	MacAddress string `json:"macAddress"`
}

func New(config []byte) (*VirtualizationFramework, error) {
	hv := &VirtualizationFramework{
		vms: make(map[string]virtualMachine),
	}

	if len(config) > 0 {
		if err := json.Unmarshal(config, &hv.cfg); err != nil {
			return nil, fmt.Errorf("invalid config: %w", err)
		}
	}

	if hv.cfg.ImageDirectory == "" || hv.cfg.WorkingDirectory == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("unable to get current user directory: %w", err)
		}

		if hv.cfg.ImageDirectory == "" {
			hv.cfg.ImageDirectory = filepath.Join(home, ".nesting/images")
			os.MkdirAll(hv.cfg.ImageDirectory, 0o777)
		}
		if hv.cfg.WorkingDirectory == "" {
			hv.cfg.WorkingDirectory = filepath.Join(home, ".nesting/data")
			os.MkdirAll(hv.cfg.WorkingDirectory, 0o777)
		}
	}

	return hv, nil
}

func (hv *VirtualizationFramework) Init(ctx context.Context, config []byte) error {
	if len(config) > 0 {
		if err := json.Unmarshal(config, &hv.cfg); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}
	}

	return nil
}

func (hv *VirtualizationFramework) Shutdown(ctx context.Context) error {
	return nil
}

func (hv *VirtualizationFramework) Create(ctx context.Context, name string) (vm hypervisor.VirtualMachine, err error) {
	cfg, err := hv.cloneVM(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("cloning vm: %w", err)
	}

	hardwareModel, err := vz.NewMacHardwareModelWithData(cfg.HardwareModel)
	if err != nil {
		return nil, fmt.Errorf("creating hardware model: %w", err)
	}

	auxStorage, err := vz.NewMacAuxiliaryStorage(
		filepath.Join(hv.cfg.WorkingDirectory, cfg.id, "nvram.bin"),
		vz.WithCreatingMacAuxiliaryStorage(hardwareModel),
	)
	if err != nil {
		return nil, fmt.Errorf("creating auxiliary storage: %w", err)
	}

	machineIdentifier, err := vz.NewMacMachineIdentifierWithData(cfg.ECID)
	if err != nil {
		return nil, fmt.Errorf("creating machine identifier: %w", err)
	}

	var bootloader vz.BootLoader
	if cfg.OS == "darwin" {
		bootloader, err = vz.NewMacOSBootLoader()
	} else {
		bootloader, err = vz.NewEFIBootLoader()
	}
	if err != nil {
		return nil, fmt.Errorf("creating bootloader: %w", err)
	}

	vzVMCfg, err := vz.NewVirtualMachineConfiguration(
		bootloader,
		cfg.CPUCount,
		cfg.MemorySize,
	)
	if err != nil {
		return nil, fmt.Errorf("creating virtual machine configuration: %w", err)
	}

	platformCfg, err := vz.NewMacPlatformConfiguration(
		vz.WithMacAuxiliaryStorage(auxStorage),
		vz.WithMacHardwareModel(hardwareModel),
		vz.WithMacMachineIdentifier(machineIdentifier),
	)
	if err != nil {
		return nil, fmt.Errorf("creating platform configuration: %w", err)
	}

	vzVMCfg.SetPlatformVirtualMachineConfiguration(platformCfg)

	diskImageAttachment, err := vz.NewDiskImageStorageDeviceAttachmentWithCacheAndSync(
		filepath.Join(hv.cfg.WorkingDirectory, cfg.id, "disk.img"),
		false,
		vz.DiskImageCachingModeAutomatic,
		vz.DiskImageSynchronizationModeNone,
	)
	if err != nil {
		return nil, fmt.Errorf("creating disk attachment: %w", err)
	}

	blockDeviceConfig, err := vz.NewVirtioBlockDeviceConfiguration(diskImageAttachment)
	if err != nil {
		return nil, fmt.Errorf("creating block device configuration for disk: %w", err)
	}

	vzVMCfg.SetStorageDevicesVirtualMachineConfiguration([]vz.StorageDeviceConfiguration{blockDeviceConfig})

	socketDeviceCfg, err := vz.NewVirtioSocketDeviceConfiguration()
	if err != nil {
		return nil, fmt.Errorf("creating socket device configuration: %w", err)
	}

	vzVMCfg.SetSocketDevicesVirtualMachineConfiguration([]vz.SocketDeviceConfiguration{socketDeviceCfg})

	networkDeviceConfig, cleanup, addr, err := createNetworkDeviceConfiguration()
	if err != nil {
		return nil, fmt.Errorf("creating network device config: %w", err)
	}
	vzVMCfg.SetNetworkDevicesVirtualMachineConfiguration([]*vz.VirtioNetworkDeviceConfiguration{
		networkDeviceConfig,
	})

	vzvm, err := vz.NewVirtualMachine(vzVMCfg)
	if err != nil {
		return nil, fmt.Errorf("creating vm: %w", err)
	}

	if err := vzvm.Start(); err != nil {
		return nil, fmt.Errorf("starting vm: %w", err)
	}

	wg, ctx := errgroup.WithContext(context.Background())

	running := make(chan struct{})
	wg.Go(func() error {
		defer cleanup()

		for state := range vzvm.StateChangedNotify() {
			switch state {
			case vz.VirtualMachineStateRunning:
				close(running)

			case vz.VirtualMachineStateError:
				return fmt.Errorf("internal VM error")

			case vz.VirtualMachineStateStopped:
				return errVirtualMachineStopped
			}
		}

		return nil
	})

	// wait for vm to start running or exit with a startup error
	select {
	case <-running:
	case <-ctx.Done():
		return nil, wg.Wait()
	}

	hv.mu.Lock()
	hv.vms[cfg.id] = virtualMachine{
		id:       cfg.id,
		addr:     addr,
		name:     name,
		vm:       vzvm,
		shutdown: wg.Wait,
	}
	hv.mu.Unlock()

	return hypervisor.VirtualMachineInfo{
		Id:   cfg.id,
		Name: name,
		Addr: addr,
	}, nil
}

func (hv *VirtualizationFramework) Delete(ctx context.Context, id string) error {
	hv.mu.Lock()
	vm, ok := hv.vms[id]
	hv.mu.Unlock()

	if !ok {
		return fmt.Errorf("no vm (%v) found", id)
	}

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// force stop, if we can
		if vm.vm.CanStop() {
			if err := vm.vm.Stop(); err != nil {
				return fmt.Errorf("stopping vm: %w", err)
			}
			break
		}

		// request stop, if we can
		if vm.vm.CanRequestStop() {
			ok, err := vm.vm.RequestStop()
			if err != nil {
				return fmt.Errorf("request stopping vm: %w", err)
			}

			if ok {
				break
			}
		}

		time.Sleep(time.Second)
	}

	// wait for shutdown
	vm.shutdown()

	if err := os.RemoveAll(filepath.Join(hv.cfg.WorkingDirectory, id)); err != nil {
		return fmt.Errorf("deleting vm dir: %w", err)
	}

	hv.mu.Lock()
	delete(hv.vms, id)
	hv.mu.Unlock()

	return nil
}

func (hv *VirtualizationFramework) List(ctx context.Context) ([]hypervisor.VirtualMachine, error) {
	hv.mu.Lock()
	defer hv.mu.Unlock()

	vms := make([]hypervisor.VirtualMachine, 0, len(hv.vms))
	for _, vm := range hv.vms {
		vms = append(vms, hypervisor.VirtualMachineInfo{
			Id:   vm.id,
			Name: vm.name,
			Addr: vm.addr,
		})
	}

	return vms, nil
}
