package control

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type CreateOptions struct {
	Id      string
	Name    string
	Timeout time.Duration
}

func VirtualMachineCreate(ctx context.Context, opts CreateOptions) (func(), error) {
	if _, err := run(ctx, "clone", opts.Name, opts.Id); err != nil {
		return func() {}, fmt.Errorf("cloning image %s (%s): %w", opts.Id, opts.Name, err)
	}

	errCh := make(chan error, 2)

	dctx, cancel := context.WithCancel(context.Background())
	go func() {
		_, err := run(dctx, "run", opts.Id, "--no-graphics", "--with-softnet")
		errCh <- err
	}()

	go func() {
		_, err := VirtualMachineAddress(ctx, opts.Id, opts.Timeout)
		errCh <- err
	}()

	// wait for either the run or ip command to exit
	err := <-errCh
	if err != nil {
		cancel()
	}

	return cancel, err
}

func VirtualMachineDelete(ctx context.Context, name string) error {
	if _, err := run(ctx, "delete", name); err != nil {
		return fmt.Errorf("deleting image: %w", err)
	}

	return nil
}

func VirtualMachineAddress(ctx context.Context, name string, timeout time.Duration) (string, error) {
	ip, err := run(ctx, "ip", name, "--wait", strconv.Itoa(int(timeout.Seconds())))
	if err != nil {
		return "", fmt.Errorf("fetching address: %w", err)
	}

	return strings.TrimSpace(ip), nil
}

func VirtualMachineList(ctx context.Context, prefix string) ([]string, error) {
	rawList, err := run(ctx, "list")
	if err != nil {
		return nil, err
	}

	var sourceIdx int
	var nameIdx int
	var fields int
	var names []string
	for _, line := range strings.Split(rawList, "\n") {
		record := strings.FieldsFunc(line, func(r rune) bool {
			return r == '\t'
		})

		// parse header
		if fields == 0 {
			fields = len(record)
			for idx, header := range record {
				switch strings.ToLower(header) {
				case "source":
					sourceIdx = idx
				case "name":
					nameIdx = idx
				}
			}
			continue
		}

		// parse record
		if fields != len(record) {
			continue
		}

		if record[sourceIdx] != "local" {
			continue
		}

		if !strings.HasPrefix(record[nameIdx], prefix) {
			continue
		}

		names = append(names, record[nameIdx])
	}

	return names, nil
}

func run(ctx context.Context, commands ...string) (string, error) {
	var stdout strings.Builder
	var stderr strings.Builder

	cmd := exec.CommandContext(ctx, "tart", commands...)
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
