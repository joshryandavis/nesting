//go:build darwin && arm64

package virtualizationframework

import (
	"fmt"
	"net"
	"sync"

	vmnet "github.com/Code-Hex/gvisor-vmnet"
	"github.com/Code-Hex/vz/v3"
)

var portMu sync.Mutex

func newLinkDevice(network *vmnet.Network) (*vmnet.LinkDevice, *vz.MACAddress, int, error) {
	mac, err := vz.NewRandomLocallyAdministeredMACAddress()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("creating mac address: %v", err)
	}

	portMu.Lock()
	defer portMu.Unlock()

	// find a random free local port we can listen on
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, nil, 0, fmt.Errorf("finding free local port: %w", err)
	}

	// once we have the port, we can immediately close the listener and pass the
	// port along to the new link device, which will then create a new listener on it.
	//
	// this does mean there's a very slim chance that another process could "steal" this port
	// before we begin listening on it again, but it won't be this process, because we use
	// the portMu lock above.
	//
	// in the event that a port is already being listened on, this will return an error and
	// the network won't be created.
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	dev, err := network.NewLinkDevice(mac.HardwareAddr(),
		vmnet.WithTCPIncomingForward(port, 22),
		vmnet.WithSendBufferSize(1024*1024),
	)

	return dev, mac, port, err
}

func createNetworkDeviceConfiguration() (*vz.VirtioNetworkDeviceConfiguration, func(), string, error) {
	network, err := vmnet.New("192.168.127.0/24")
	if err != nil {
		return nil, nil, "", fmt.Errorf("creating network: %w", err)
	}

	dev, mac, port, err := newLinkDevice(network)
	if err != nil {
		network.Shutdown()
		return nil, nil, "", fmt.Errorf("creating link device: %w", err)
	}

	cleanup := func() {
		dev.Close()
		network.Shutdown()
	}

	attachment, err := vz.NewFileHandleNetworkDeviceAttachment(dev.File())
	if err != nil {
		cleanup()
		return nil, nil, "", fmt.Errorf("creating file handle network device attachment: %w", err)
	}

	config, err := vz.NewVirtioNetworkDeviceConfiguration(attachment)
	if err != nil {
		cleanup()
		return nil, nil, "", fmt.Errorf("creating virtio netwqork device config: %w", err)
	}

	config.SetMACAddress(mac)

	return config, cleanup, fmt.Sprintf("127.0.0.1:%d", port), nil
}
