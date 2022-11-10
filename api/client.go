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

type Client struct {
	conn   *grpc.ClientConn
	client proto.NestingClient
}

type Dialer func(ctx context.Context, network, address string) (net.Conn, error)

func New(client *grpc.ClientConn) *Client {
	return &Client{
		conn:   client,
		client: proto.NewNestingClient(client),
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

func (c *Client) Init(ctx context.Context, config []byte) error {
	_, err := c.client.Init(ctx, &proto.InitRequest{
		Config: config,
	})

	return err
}

func (c *Client) Shutdown(ctx context.Context) error {
	_, err := c.client.Shutdown(ctx, &proto.ShutdownRequest{})

	return err
}

func (c *Client) Create(ctx context.Context, name string) (hypervisor.VirtualMachine, error) {
	return c.client.Create(ctx, &proto.CreateRequest{
		Name: name,
	})
}

func (c *Client) Delete(ctx context.Context, id string) error {
	_, err := c.client.Delete(ctx, &proto.DeleteRequest{
		Id: id,
	})

	return err
}

func (c *Client) List(ctx context.Context) ([]hypervisor.VirtualMachine, error) {
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

func (c *Client) Close() error {
	return c.conn.Close()
}
