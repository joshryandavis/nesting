package parallels

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitlab.com/ajwalker/nesting/hypervisor/parallels/internal/control"
)

var ErrNoNetworkAvailable = errors.New("no network available")

type Network string

func (p *Parallels) getNetwork() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for network, acquired := range p.networks {
		if !acquired {
			p.networks[network] = true
			return network, nil
		}
	}

	return "", ErrNoNetworkAvailable
}

func (p *Parallels) putNetwork(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.networks[name] = false
}

func (p *Parallels) populateNetworks(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.networks == nil {
		networks, err := control.NetworkList(ctx, networkNamePrefix)
		if err != nil {
			return fmt.Errorf("fetching initial network list: %v", err)
		}

		p.networks = make(map[string]bool, len(networks))
		for _, network := range networks {
			p.networks[network] = false
		}
	}

	return nil
}

// getAddress returns the IP address of a VM via its MAC.
//
// If a timeout is provided, this call blocks until the lease file exists and
// returns the address or the timeout is exceeded.
//
// If no timeout is provided, it returns immediately with or without an address.
// The lease not existing is not treated as an error.
func getAddress(ctx context.Context, mac string, timeout time.Duration) (string, error) {
	if timeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	for {
		buf, err := os.ReadFile(filepath.Join("/tmp/parallels.leases/", strings.ToLower(mac)))
		if err == nil {
			return net.IP(buf).String(), nil
		}

		if errors.Is(err, os.ErrNotExist) {
			if timeout == 0 {
				return "", nil
			}

			if ctx.Err() != nil {
				return "", ctx.Err()
			}

			time.Sleep(time.Second)
			continue
		}

		return "", err
	}
}

func removeLease(mac string) {
	os.RemoveAll(filepath.Join("/tmp/parallels.leases/", strings.ToLower(mac)))
}
