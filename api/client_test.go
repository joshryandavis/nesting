package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/gitlab-org/fleeting/nesting/api/internal/proto"
	"gitlab.com/gitlab-org/fleeting/nesting/api/internal/proto/mocks"
	"gitlab.com/gitlab-org/fleeting/nesting/hypervisor"
)

func TestCreate(t *testing.T) {
	testCases := map[string]struct {
		name            string
		slot            *int32
		expect          clientExpectation
		wantVm          hypervisor.VirtualMachine
		wantStompedVmId *string
		wantErr         bool
	}{
		"success": {
			name: "name",
			expect: clientCreate(&proto.CreateRequest{Name: "name"},
				&proto.CreateResponse{Vm: &proto.VirtualMachine{Name: "name"}}, nil),
			wantVm: &hypervisor.VirtualMachineInfo{Name: "name"},
		},
		"error": {
			name: "name",
			expect: clientCreate(&proto.CreateRequest{Name: "name"},
				nil, fmt.Errorf("no can do")),
			wantErr: true,
		},
		"with a slot": {
			name: "name",
			slot: int32Ref(0),
			expect: clientCreate(&proto.CreateRequest{Name: "name", Slot: int32Ref(0)},
				&proto.CreateResponse{Vm: &proto.VirtualMachine{Name: "name"}}, nil),
			wantVm: &hypervisor.VirtualMachineInfo{Name: "name"},
		},
		"with a slot stomp": {
			name: "name",
			slot: int32Ref(0),
			expect: clientCreate(&proto.CreateRequest{Name: "name", Slot: int32Ref(0)},
				&proto.CreateResponse{Vm: &proto.VirtualMachine{Name: "name"}, StompedVmId: stringRef("abc")}, nil),
			wantVm:          &hypervisor.VirtualMachineInfo{Name: "name"},
			wantStompedVmId: stringRef("abc"),
		},
		"nil response": {
			name: "name",
			expect: clientCreate(&proto.CreateRequest{Name: "name"},
				nil, nil),
			wantVm:          nil,
			wantStompedVmId: nil,
			wantErr:         false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			m := mocks.NewNestingClient(t)
			tc.expect(m)
			c := &client{
				client: m,
			}
			vm, stompedVmId, err := c.Create(context.TODO(), tc.name, tc.slot)
			assertHypervisorVmEqual(t, tc.wantVm, vm)
			assert.Equal(t, tc.wantStompedVmId, stompedVmId)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func assertHypervisorVmEqual(t *testing.T, want, got hypervisor.VirtualMachine) {
	if want == nil && got == nil {
		return
	}
	if want == nil {
		t.Errorf("want nil. got %v", got)
		return
	}
	if got == nil {
		t.Errorf("want %v. got nil", want)
	}
	for _, f := range [][2]func() string{
		{want.GetId, got.GetId},
		{want.GetName, got.GetName},
		{want.GetAddr, got.GetAddr},
	} {
		assert.Equal(t, f[0](), f[1]())
	}
}

type clientExpectation func(m *mocks.NestingClient)

func clientCreate(request *proto.CreateRequest, response *proto.CreateResponse, err error) clientExpectation {
	return func(m *mocks.NestingClient) {
		m.EXPECT().Create(context.TODO(), request).Return(response, err)
	}
}
