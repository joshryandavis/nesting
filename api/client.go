package api

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"gitlab.com/ajwalker/nesting/api/internal/proto"
	"gitlab.com/ajwalker/nesting/hypervisor"
)

type Client struct {
	client proto.NestingClient
}

func New(client *grpc.ClientConn) *Client {
	return &Client{
		client: proto.NewNestingClient(client),
	}
}

func DefaultConn() (*grpc.ClientConn, error) {
	return grpc.Dial(
		"unix",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(c context.Context, s string) (net.Conn, error) {
			return net.Dial("unix", socketPath())
		}),
	)
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
