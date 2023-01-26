package api

import (
	"context"
	"net"
	"net/url"
	"path/filepath"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"gitlab.com/gitlab-org/fleeting/nesting/api/internal/proto"
	"gitlab.com/gitlab-org/fleeting/nesting/hypervisor"
)

//go:generate mockery --name=Client --with-expecter
type Client interface {
	Init(ctx context.Context, config []byte) error
	Shutdown(ctx context.Context) error
	Create(ctx context.Context, name string, slot *int32) (vm hypervisor.VirtualMachine, stompedVmId *string, err error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]hypervisor.VirtualMachine, error)
	Close() error
}

var _ Client = &client{}

type client struct {
	conn   *grpc.ClientConn
	client proto.NestingClient
}

type Dialer func(ctx context.Context, network, address string) (net.Conn, error)

func New(c *grpc.ClientConn) Client {
	return &client{
		conn:   c,
		client: proto.NewNestingClient(c),
	}
}

func DefaultConn() (*grpc.ClientConn, error) {
	return NewClientConn("", nil)
}

func NewClientConn(target string, dialer Dialer) (*grpc.ClientConn, error) {
	if target == "" {
		target = socketPath()
		if filepath.IsAbs(target) {
			target = "unix://" + target
		} else {
			target = "unix:" + target
		}
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if dialer != nil {
		opts = append(opts, grpc.WithContextDialer(func(c context.Context, s string) (net.Conn, error) {
			network, address := parseDialTarget(s)

			return dialer(c, network, address)
		}))
	}

	return grpc.Dial(target, opts...)
}

func parseDialTarget(target string) (string, string) {
	network := "tcp"

	// unix://absolute
	if strings.Contains(target, ":/") {
		uri, err := url.Parse(target)
		if err != nil {
			return network, target
		}

		if uri.Path == "" {
			return uri.Scheme, uri.Host
		}
		return uri.Scheme, uri.Path
	}

	// unix:relative-path
	if network, path, found := strings.Cut(target, ":"); found {
		return network, path
	}

	// tcp://target
	return network, target
}

func (c *client) Init(ctx context.Context, config []byte) error {
	_, err := c.client.Init(ctx, &proto.InitRequest{
		Config: config,
	})

	return err
}

func (c *client) Shutdown(ctx context.Context) error {
	_, err := c.client.Shutdown(ctx, &proto.ShutdownRequest{})

	return err
}

func (c *client) Create(ctx context.Context, name string, slot *int32) (vm hypervisor.VirtualMachine, stompedVmId *string, err error) {
	response, err := c.client.Create(ctx, &proto.CreateRequest{
		Name: name,
		Slot: slot,
	})
	if err != nil {
		return nil, nil, err
	}
	if response == nil {
		return nil, nil, nil
	}
	return response.Vm, response.StompedVmId, nil
}

func (c *client) Delete(ctx context.Context, id string) error {
	_, err := c.client.Delete(ctx, &proto.DeleteRequest{
		Id: id,
	})

	return err
}

func (c *client) List(ctx context.Context) ([]hypervisor.VirtualMachine, error) {
	results, err := c.client.List(ctx, &proto.ListRequest{})
	if err != nil {
		return nil, err
	}

	vms := make([]hypervisor.VirtualMachine, 0, len(results.Vms))
	for _, vm := range results.Vms {
		vms = append(vms, vm)
	}

	return vms, nil
}

func (c *client) Close() error {
	return c.conn.Close()
}
