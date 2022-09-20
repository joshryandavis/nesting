package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"google.golang.org/grpc"

	"gitlab.com/ajwalker/nesting/api/internal/proto"
	"gitlab.com/ajwalker/nesting/hypervisor"
)

var (
	errAlreadyInitialized = errors.New("already initialized")
	errNotInitialized     = errors.New("not initialized")
)

type server struct {
	hv     hypervisor.Hypervisor
	mu     sync.Mutex
	inited bool

	proto.UnimplementedNestingServer
}

func (s *server) Init(ctx context.Context, req *proto.InitRequest) (*proto.InitResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.inited {
		return nil, errAlreadyInitialized
	}

	err := s.hv.Init(ctx, req.Config)
	if err == nil {
		s.inited = true
	}

	return &proto.InitResponse{}, err
}

func (s *server) Shutdown(ctx context.Context, _ *proto.ShutdownRequest) (*proto.ShutdownResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.inited {
		return nil, errNotInitialized
	}

	err := s.hv.Shutdown(ctx)
	if err == nil {
		s.inited = false
	}

	return &proto.ShutdownResponse{}, err
}

func (s *server) Create(ctx context.Context, req *proto.CreateRequest) (*proto.VirtualMachine, error) {
	if !s.initialized() {
		return nil, errNotInitialized
	}

	vm, err := s.hv.Create(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	return &proto.VirtualMachine{
		Id:   vm.GetId(),
		Name: vm.GetName(),
		Addr: vm.GetAddr(),
	}, nil
}

func (s *server) Delete(ctx context.Context, req *proto.DeleteRequest) (*proto.DeleteResponse, error) {
	if !s.initialized() {
		return nil, errNotInitialized
	}

	return &proto.DeleteResponse{}, s.hv.Delete(ctx, req.Id)
}

func (s *server) List(ctx context.Context, req *proto.ListRequest) (*proto.ListResponse, error) {
	if !s.initialized() {
		return nil, errNotInitialized
	}

	vms, err := s.hv.List(ctx)

	var list proto.ListResponse
	for _, vm := range vms {
		list.Vms = append(list.Vms, &proto.VirtualMachine{
			Id:   vm.GetId(),
			Name: vm.GetName(),
			Addr: vm.GetAddr(),
		})
	}

	return &list, err
}

func (s *server) initialized() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.inited
}

func Serve(ctx context.Context, hv hypervisor.Hypervisor) error {
	socket := socketPath()
	os.MkdirAll(filepath.Dir(socket), 0777)

	listener, err := net.Listen("unix", socket)
	if err != nil {
		return fmt.Errorf("creating listener: %w", err)
	}
	defer os.RemoveAll(socket)
	defer listener.Close()

	srv := grpc.NewServer()
	proto.RegisterNestingServer(srv, &server{hv: hv})

	// the service being shutdown also calls Shutdown on the hypervisor impl
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		hv.Shutdown(ctx)
	}()

	defer srv.Stop()
	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()

	return srv.Serve(listener)
}

func socketPath() string {
	name := os.Getenv("NESTING_SOCKET")
	if name != "" {
		return name
	}

	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		name = filepath.Join(home, "Library/Application Support")
	case "windows":
		name = os.Getenv("LOCALAPPDATA")
	default:
		name = os.Getenv("XDG_RUNTIME_DIR")
	}
	if name == "" {
		name = home
	}

	return filepath.Join(name, "nesting.sock")
}
