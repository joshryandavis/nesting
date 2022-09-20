package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	controlCmd       = "prlctl"
	serverControlCmd = "prlsrvctl"
)

var (
	errLicenseInstallFailure = errors.New("failed to install license")
	errLicenseRemoveFailure  = errors.New("failed to remove license")
)

type vmListItem struct {
	Name        string
	Description string
	Hardware    struct {
		Net0 struct {
			Iface string `json:"iface"`
			Mac   string `json:"mac"`
		} `json:"net0"`
	}
}

func InstallLicense(ctx context.Context, key string) error {
	_, err := run(ctx, serverControlCmd, "install-license", "-k", key)
	if err != nil {
		return errLicenseInstallFailure
	}

	return nil
}

func RemoveLicense(ctx context.Context) error {
	_, err := run(ctx, serverControlCmd, "deactivate-license")
	if err != nil {
		// don't treat an error about not being activated as an error
		if strings.Contains(err.Error(), "deactivate the license") ||
			strings.Contains(err.Error(), "product could not be activated") {
			return nil
		}
		return errLicenseRemoveFailure
	}

	return nil
}

type CreateOptions struct {
	Id         string
	ImagePath  string
	WorkingDir string
	MAC        string
	Network    string
}

func VirtualMachineCreate(ctx context.Context, opts CreateOptions) error {
	name := filepath.Base(opts.ImagePath)
	name = strings.TrimSuffix(name, ".pvm")

	run(ctx, controlCmd, "register", opts.ImagePath)

	if _, err := run(ctx, controlCmd, "clone", name, "--name", opts.Id, "--linked", "--dst", opts.WorkingDir); err != nil {
		return fmt.Errorf("cloning image %s (%s): %w", opts.Id, name, err)
	}

	if _, err := run(ctx, controlCmd, "set", opts.Id, "--description", name); err != nil {
		return fmt.Errorf("updating image settings %s: %w", opts.Id, err)
	}

	if _, err := run(ctx, controlCmd, "set", opts.Id, "--device-set", "net0", "--type", "host-only", "--iface", opts.Network, "--adapter-type", "virtio", "--mac", opts.MAC); err != nil {
		return fmt.Errorf("updating image settings %s: %w", opts.Id, err)
	}

	if _, err := run(ctx, controlCmd, "start", opts.Id); err != nil {
		return fmt.Errorf("starting image: %w", err)
	}

	return nil
}

func VirtualMachineDelete(ctx context.Context, name string) error {
	if _, err := run(ctx, controlCmd, "stop", name, "--kill"); err != nil {
		return fmt.Errorf("deleting image: %w", err)
	}

	if _, err := run(ctx, controlCmd, "delete", name); err != nil {
		return fmt.Errorf("deleting image: %w", err)
	}

	return nil
}

func VirtualMachineList(ctx context.Context, prefix string) ([]vmListItem, error) {
	rawList, err := run(ctx, controlCmd, "list", "-a", "-i", "-j")
	if err != nil {
		return nil, err
	}

	var items []vmListItem
	if err := json.Unmarshal([]byte(rawList), &items); err != nil {
		return nil, err
	}

	filtered := make([]vmListItem, 0, len(items))
	for _, item := range items {
		if !strings.HasPrefix(item.Name, prefix) {
			continue
		}

		filtered = append(filtered, item)
	}

	return filtered, nil
}

func NetworkList(ctx context.Context, prefix string) ([]string, error) {
	rawList, err := run(ctx, serverControlCmd, "net", "list", "-j")
	if err != nil {
		return nil, err
	}

	var list []struct {
		NetworkID string `json:"Network ID"`
	}

	if err := json.Unmarshal([]byte(rawList), &list); err != nil {
		return nil, err
	}

	var networks []string
	for _, item := range list {
		if !strings.HasPrefix(item.NetworkID, prefix) {
			continue
		}
		networks = append(networks, item.NetworkID)
	}

	return networks, nil
}

func run(ctx context.Context, commands ...string) (string, error) {
	var stdout strings.Builder
	var stderr strings.Builder

	cmd := exec.CommandContext(ctx, commands[0], commands[1:]...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	var errExit *exec.ExitError
	if errors.As(err, &errExit) {
		return stdout.String(), fmt.Errorf("%s: %w (%s)", strings.Join(commands, " "), err, stderr.String())
	}
	if err != nil {
		return stdout.String(), fmt.Errorf("%s: %w", commands[0], err)
	}

	return stdout.String(), nil
}
