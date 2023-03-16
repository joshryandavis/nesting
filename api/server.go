package api

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.com/gitlab-org/fleeting/nesting/api/internal/proto"
	"gitlab.com/gitlab-org/fleeting/nesting/hypervisor"
)

var (
	ErrAlreadyInitialized = status.Error(codes.FailedPrecondition, "already initialized")
	ErrNotInitialized     = status.Error(codes.FailedPrecondition, "not initialized")
)

type server struct {
	hv     hypervisor.Hypervisor
	mu     sync.Mutex
	inited bool
	slots  map[int32]string

	proto.UnimplementedNestingServer
}

func (s *server) Init(ctx context.Context, req *proto.InitRequest) (*proto.InitResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.inited {
		return nil, ErrAlreadyInitialized
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
		return nil, ErrAlreadyInitialized
	}

	err := s.hv.Shutdown(ctx)
	if err == nil {
		s.inited = false
	}

	return &proto.ShutdownResponse{}, err
}

func (s *server) Create(ctx context.Context, req *proto.CreateRequest) (*proto.CreateResponse, error) {
	if !s.initialized() {
		return nil, ErrNotInitialized
	}

	slotsInUse := req.Slot != nil
	var stompedVmId *string
	if slotsInUse {
		id, err := s.clearSlot(ctx, *req.Slot)
		if err != nil {
			return nil, err
		}
		stompedVmId = id
	}

	vm, err := s.hv.Create(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if slotsInUse {
		s.slots[*req.Slot] = vm.GetId()
	}

	return &proto.CreateResponse{
		Vm: &proto.VirtualMachine{
			Id:   vm.GetId(),
			Name: vm.GetName(),
			Addr: vm.GetAddr(),
		},
		StompedVmId: stompedVmId,
	}, nil
}

func (s *server) Delete(ctx context.Context, req *proto.DeleteRequest) (*proto.DeleteResponse, error) {
	if !s.initialized() {
		return nil, ErrNotInitialized
	}

	err := s.hv.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for slot, id := range s.slots {
		if id == req.Id {
			delete(s.slots, slot)
		}
	}

	return &proto.DeleteResponse{}, nil
}

func (s *server) List(ctx context.Context, req *proto.ListRequest) (*proto.ListResponse, error) {
	if !s.initialized() {
		return nil, ErrNotInitialized
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

func (s *server) clearSlot(ctx context.Context, slot int32) (*string, error) {
	s.mu.Lock()
	id, ok := s.slots[slot]
	s.mu.Unlock()

	if !ok {
		return nil, nil
	}

	if _, err := s.Delete(ctx, &proto.DeleteRequest{Id: id}); err != nil {
		return nil, fmt.Errorf("clearing slot: %w", err)
	}

	return &id, nil
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
	proto.RegisterNestingServer(srv, &server{
		hv:    hv,
		slots: make(map[int32]string),
	})

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
