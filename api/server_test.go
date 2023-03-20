package api

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"gitlab.com/gitlab-org/fleeting/nesting/api/internal/proto"
	"gitlab.com/gitlab-org/fleeting/nesting/hypervisor"
	"gitlab.com/gitlab-org/fleeting/nesting/hypervisor/mocks"
)

func TestServer(t *testing.T) {

	testCases := map[string][]struct {
		request  interface{}
		expect   []expectation
		response interface{}
		err      bool
	}{
		"double init": {{ // what does it mean?
			request: &proto.InitRequest{Config: []byte{}},
			expect: []expectation{
				hvInit([]byte{}, nil),
			},
			response: &proto.InitResponse{},
		}, {
			request: &proto.InitRequest{Config: []byte{}},
			err:     true,
		}},

		"full cycle of init, create, delete, list and shutdown": {{
			request: &proto.InitRequest{Config: []byte{}},
			expect: []expectation{
				hvInit([]byte{}, nil),
			},
			response: &proto.InitResponse{},
		}, {
			request: &proto.CreateRequest{Name: "name-1"},
			expect: []expectation{
				hvCreate("name-1", hypervisor.VirtualMachineInfo{Name: "name-1", Id: "id-1", Addr: "1.1.1.1"}, nil),
			},
			response: &proto.CreateResponse{Vm: &proto.VirtualMachine{Name: "name-1", Id: "id-1", Addr: "1.1.1.1"}},
		}, {
			request: &proto.DeleteRequest{Id: "id-1"},
			expect: []expectation{
				hvDelete("id-1", nil),
			},
			response: &proto.DeleteResponse{},
		}, {
			request: &proto.ListRequest{},
			expect: []expectation{
				hvList([]hypervisor.VirtualMachineInfo{{Name: "name-1", Id: "id-1", Addr: "1.1.1.1"}}, nil),
			},
			response: &proto.ListResponse{Vms: []*proto.VirtualMachine{{Name: "name-1", Id: "id-1", Addr: "1.1.1.1"}}},
		}, {
			request: &proto.ShutdownRequest{},
			expect: []expectation{
				hvShutdown(nil),
			},
			response: &proto.ShutdownResponse{},
		}},

		"two non-slot creates": {{
			request: &proto.InitRequest{Config: []byte{}},
			expect: []expectation{
				hvInit([]byte{}, nil),
			},
			response: &proto.InitResponse{},
		}, {
			request: &proto.CreateRequest{Name: "name-1", Slot: nil}, // first vm (no slot)
			expect: []expectation{
				hvCreate("name-1", hypervisor.VirtualMachineInfo{Name: "name-1", Id: "id-1", Addr: "1.1.1.1"}, nil),
			},
			response: &proto.CreateResponse{Vm: &proto.VirtualMachine{Name: "name-1", Id: "id-1", Addr: "1.1.1.1"}},
		}, {
			request: &proto.CreateRequest{Name: "name-2", Slot: nil}, // second vm (no slot)
			expect: []expectation{
				hvCreate("name-2", hypervisor.VirtualMachineInfo{Name: "name-2", Id: "id-2", Addr: "2.2.2.2"}, nil),
			},
			response: &proto.CreateResponse{Vm: &proto.VirtualMachine{Name: "name-2", Id: "id-2", Addr: "2.2.2.2"}},
		}},

		"creation in different slots": {{
			request: &proto.InitRequest{Config: []byte{}},
			expect: []expectation{
				hvInit([]byte{}, nil),
			},
			response: &proto.InitResponse{},
		}, {
			request: &proto.CreateRequest{Name: "name-1", Slot: int32Ref(0)}, // vm in slot 0
			expect: []expectation{
				hvCreate("name-1", hypervisor.VirtualMachineInfo{Name: "name-1", Id: "id-1", Addr: "1.1.1.1"}, nil),
			},
			response: &proto.CreateResponse{Vm: &proto.VirtualMachine{Name: "name-1", Id: "id-1", Addr: "1.1.1.1"}},
		}, {
			request: &proto.CreateRequest{Name: "name-2", Slot: int32Ref(1)}, // vm in slot 1
			expect: []expectation{
				// vm in slot 1 doesn't stomp vm in slot 0
				hvCreate("name-2", hypervisor.VirtualMachineInfo{Name: "name-2", Id: "id-2", Addr: "2.2.2.2"}, nil),
			},
			response: &proto.CreateResponse{Vm: &proto.VirtualMachine{Name: "name-2", Id: "id-2", Addr: "2.2.2.2"}},
		}},

		"slot stomp": {{
			request: &proto.InitRequest{Config: []byte{}},
			expect: []expectation{
				hvInit([]byte{}, nil),
			},
			response: &proto.InitResponse{},
		}, {
			request: &proto.CreateRequest{Name: "name-1", Slot: int32Ref(0)}, // first vm in slot 0
			expect: []expectation{
				hvCreate("name-1", hypervisor.VirtualMachineInfo{Name: "name-1", Id: "id-1", Addr: "1.1.1.1"}, nil),
			},
			response: &proto.CreateResponse{Vm: &proto.VirtualMachine{Name: "name-1", Id: "id-1", Addr: "1.1.1.1"}},
		}, {
			request: &proto.CreateRequest{Name: "name-2", Slot: int32Ref(0)}, // second vm in slot 0
			expect: []expectation{
				hvDelete("id-1", nil), // clear slot 0
				hvCreate("name-2", hypervisor.VirtualMachineInfo{Name: "name-2", Id: "id-2", Addr: "2.2.2.2"}, nil),
			},
			response: &proto.CreateResponse{
				Vm:          &proto.VirtualMachine{Name: "name-2", Id: "id-2", Addr: "2.2.2.2"},
				StompedVmId: stringRef("id-1"),
			},
		}},

		"slot reuse": {{
			request: &proto.InitRequest{Config: []byte{}},
			expect: []expectation{
				hvInit([]byte{}, nil),
			},
			response: &proto.InitResponse{},
		}, {
			request: &proto.CreateRequest{Name: "name-1", Slot: int32Ref(0)}, // first vm in slot 0
			expect: []expectation{
				hvCreate("name-1", hypervisor.VirtualMachineInfo{Name: "name-1", Id: "id-1", Addr: "1.1.1.1"}, nil),
			},
			response: &proto.CreateResponse{Vm: &proto.VirtualMachine{Name: "name-1", Id: "id-1", Addr: "1.1.1.1"}},
		}, {
			request: &proto.DeleteRequest{Id: "id-1"}, // delete first vm
			expect: []expectation{
				hvDelete("id-1", nil),
			},
			response: &proto.DeleteResponse{},
		}, {
			request: &proto.CreateRequest{Name: "name-2", Slot: int32Ref(0)}, // second vm in slot 0
			expect: []expectation{
				hvCreate("name-2", hypervisor.VirtualMachineInfo{Name: "name-2", Id: "id-2", Addr: "2.2.2.2"}, nil),
			},
			response: &proto.CreateResponse{Vm: &proto.VirtualMachine{Name: "name-2", Id: "id-2", Addr: "2.2.2.2"}},
		}},

		// We must remember the vm id associated with a slot
		// since deletion failure means resources are still
		// allocated. This assumes that deletion is retryable.
		"don't forget about a vm when deletion fails": {{
			request: &proto.InitRequest{Config: []byte{}},
			expect: []expectation{
				hvInit([]byte{}, nil),
			},
			response: &proto.InitResponse{},
		}, {
			request: &proto.CreateRequest{Name: "name-1", Slot: int32Ref(0)},
			expect: []expectation{
				hvCreate("name-1", hypervisor.VirtualMachineInfo{Name: "name-1", Id: "id-1", Addr: "1.1.1.1"}, nil),
			},
			response: &proto.CreateResponse{Vm: &proto.VirtualMachine{Name: "name-1", Id: "id-1", Addr: "1.1.1.1"}},
		}, {
			request: &proto.DeleteRequest{Id: "id-1"},
			expect: []expectation{
				hvDelete("id-1", fmt.Errorf("no can do")),
			},
			response: &proto.DeleteResponse{},
			err:      true, // first deletion attempt fails
		}, {
			request: &proto.CreateRequest{Name: "name-2", Slot: int32Ref(0)},
			expect: []expectation{
				hvDelete("id-1", fmt.Errorf("still no can do")), // create in the same slot re-attempts deletion, fails again
			},
			err: true,
		}, {
			request: &proto.CreateRequest{Name: "name-3", Slot: int32Ref(0)},
			expect: []expectation{
				hvDelete("id-1", nil), // create in the same slot re-attempts deletion, this time succeeds
				hvCreate("name-3", hypervisor.VirtualMachineInfo{Name: "name-3", Id: "id-3", Addr: "3.3.3.3"}, nil),
			},
			response: &proto.CreateResponse{
				Vm:          &proto.VirtualMachine{Name: "name-3", Id: "id-3", Addr: "3.3.3.3"},
				StompedVmId: stringRef("id-1"),
			},
		}},

		"slot number doesn't really matter": {{
			request: &proto.InitRequest{Config: []byte{}},
			expect: []expectation{
				hvInit([]byte{}, nil),
			},
			response: &proto.InitResponse{},
		}, {
			request: &proto.CreateRequest{Name: "name-1", Slot: int32Ref(99)}, // any unique number will do
			expect: []expectation{
				hvCreate("name-1", hypervisor.VirtualMachineInfo{Name: "name-1", Id: "id-1", Addr: "1.1.1.1"}, nil),
			},
			response: &proto.CreateResponse{Vm: &proto.VirtualMachine{Name: "name-1", Id: "id-1", Addr: "1.1.1.1"}},
		}, {
			request: &proto.CreateRequest{Name: "name-2", Slot: int32Ref(99)},
			expect: []expectation{
				hvDelete("id-1", nil),
				hvCreate("name-2", hypervisor.VirtualMachineInfo{Name: "name-2", Id: "id-2", Addr: "2.2.2.2"}, nil),
			},
			response: &proto.CreateResponse{
				Vm:          &proto.VirtualMachine{Name: "name-2", Id: "id-2", Addr: "2.2.2.2"},
				StompedVmId: stringRef("id-1"),
			},
		}},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			m := mocks.NewHypervisor(t)
			s := server{
				hv:    m,
				slots: map[int32]string{},
			}
			for _, step := range testCase {
				for _, expect := range step.expect {
					expect(m)
				}
				switch r := step.request.(type) {
				case *proto.InitRequest:
					callAndAssert[*proto.InitRequest, *proto.InitResponse](t, s.Init, r, step.response, step.err)
				case *proto.CreateRequest:
					callAndAssert[*proto.CreateRequest, *proto.CreateResponse](t, s.Create, r, step.response, step.err)
				case *proto.ListRequest:
					callAndAssert[*proto.ListRequest, *proto.ListResponse](t, s.List, r, step.response, step.err)
				case *proto.DeleteRequest:
					callAndAssert[*proto.DeleteRequest, *proto.DeleteResponse](t, s.Delete, r, step.response, step.err)
				case *proto.ShutdownRequest:
					callAndAssert[*proto.ShutdownRequest, *proto.ShutdownResponse](t, s.Shutdown, r, step.response, step.err)
				default:
					t.Errorf("unhandled request type: %T", r)
				}
			}
		})
	}
}

func callAndAssert[R protoreflect.ProtoMessage, S protoreflect.ProtoMessage](t *testing.T, method func(context.Context, R) (S, error), request R, wantResponse any, wantErr bool) {
	res, err := method(context.TODO(), request)
	if wantErr {
		assert.Nil(t, res)
		assert.Error(t, err)
	} else {
		if !protobuf.Equal(res, wantResponse.(S)) { // Convert response here so the caller doesn't have to convert.
			t.Errorf("want %v. got %v", wantResponse, res)
		}
		assert.Nil(t, err)
	}
}

type expectation func(m *mocks.Hypervisor)

func hvInit(config []byte, err error) expectation {
	return func(m *mocks.Hypervisor) {
		m.EXPECT().Init(context.TODO(), config).Return(err).Once()
	}
}

func hvCreate(name string, vm hypervisor.VirtualMachine, err error) expectation {
	return func(m *mocks.Hypervisor) {
		m.EXPECT().Create(context.TODO(), name).Return(vm, err).Once()
	}
}

func hvList(vms []hypervisor.VirtualMachineInfo, err error) expectation {
	vmsI := make([]hypervisor.VirtualMachine, len(vms))
	for i, vm := range vms {
		vmsI[i] = vm
	}
	return func(m *mocks.Hypervisor) {
		m.EXPECT().List(context.TODO()).Return(vmsI, err).Once()
	}
}

func hvDelete(id string, err error) expectation {
	return func(m *mocks.Hypervisor) {
		m.EXPECT().Delete(context.TODO(), id).Return(err).Once()
	}
}

func hvShutdown(err error) expectation {
	return func(m *mocks.Hypervisor) {
		m.EXPECT().Shutdown(context.TODO()).Return(err).Once()
	}
}

// server.Serve is untested

func TestConcurrentCreateCall(t *testing.T) {
	m := mocks.NewHypervisor(t)
	s := server{
		hv:    m,
		slots: make(map[int32]string),
	}

	m.EXPECT().Init(context.TODO(), []byte{}).Return(nil).Once()

	resp, err := s.Init(context.TODO(), &proto.InitRequest{Config: []byte{}})
	require.NoError(t, err)

	t.Log("init resp", resp)

	requestsNo := 50

	runCh := make(chan struct{})
	wg := new(sync.WaitGroup)

	add := func(id int, slot *int32) {
		wg.Add(1)
		go func(wg *sync.WaitGroup, r chan struct{}, id int, slot *int32) {
			defer wg.Done()

			req := &proto.CreateRequest{Name: fmt.Sprintf("name-%d", id), Slot: slot}
			vm := hypervisor.VirtualMachineInfo{Name: fmt.Sprintf("name-%d", id), Id: fmt.Sprintf("id-%d", id), Addr: fmt.Sprintf("1.1.1.%d", id)}

			m.EXPECT().Create(context.TODO(), req.Name).Return(vm, err).Once()

			<-r
			s.Create(context.TODO(), req)
		}(wg, runCh, id, slot)
	}

	m.EXPECT().Delete(context.TODO(), mock.Anything).Return(nil)
	for i := 0; i < requestsNo; i++ {
		add(i, int32Ref(0))
	}

	close(runCh)
	wg.Wait()
}
