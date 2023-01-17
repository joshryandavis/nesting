package control

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVirtualMachineList(t *testing.T) {
	const nestingPrefix = "nesting-"
	cases := []struct {
		name   string
		expect *mockRun
		list   []string
		err    bool
	}{
		{
			name: "empty list",
			expect: &mockRun{
				commands:     []string{"list"},
				returnString: []string{},
			},
			list: []string{},
		},
		{
			name: "only headers",
			expect: &mockRun{
				commands: []string{"list"},
				returnString: []string{
					"source\tname",
				},
			},
			list: []string{},
		},
		{
			name: "one vm",
			expect: &mockRun{
				commands: []string{"list"},
				returnString: []string{
					"source\tname",
					"local\tnesting-abc",
				},
			},
			list: []string{
				"nesting-abc",
			},
		},
		{
			name: "multiple vms",
			expect: &mockRun{
				commands: []string{"list"},
				returnString: []string{
					"source\tname",
					"local\tnesting-abc",
					"local\tnesting-def",
					"local\tnesting-geh",
				},
			},
			list: []string{
				"nesting-abc",
				"nesting-def",
				"nesting-geh",
			},
		},
		{
			name: "ignore non-local images",
			expect: &mockRun{
				commands: []string{"list"},
				returnString: []string{
					"source\tname",
					"oci\tnesting-not-your-vm",
					"local\tnesting-abc",
				},
			},
			list: []string{
				"nesting-abc",
			},
		},
		{
			name: "ignore garbage",
			expect: &mockRun{
				commands: []string{"list"},
				returnString: []string{
					"source\tname",
					"local\tnesting-abc",
					"garbage!",
					"local\tnesting-def",
				},
			},
			list: []string{
				"nesting-abc",
				"nesting-def",
			},
		},
		{
			name: "ignore different prefixed vms",
			expect: &mockRun{
				commands: []string{"list"},
				returnString: []string{
					"source\tname",
					"local\tnesting-abc",
					"local\tresting-def", // resting-
					"local\ttesting-geh", // testing-
				},
			},
			list: []string{
				"nesting-abc",
			},
		},
		{
			name: "spaces in weird places",
			expect: &mockRun{
				commands: []string{"list"},
				returnString: []string{
					"source \tname ",
					"local \tnesting-abc ",
					" local\t nesting-def",
					" local \t nesting-geh ",
				},
			},
			list: []string{
				"nesting-abc",
				"nesting-def",
				"nesting-geh",
			},
		},
		{
			name: "check err",
			expect: &mockRun{
				commands:  []string{"list"},
				returnErr: fmt.Errorf("no can do"),
			},
			err: true,
		},
	}

	runFunc := run
	defer func() {
		run = runFunc
	}()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			run = tc.expect.fn()
			got, err := VirtualMachineList(context.TODO(), nestingPrefix)
			assertStringsEqual(t, tc.list, got)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
			tc.expect.verify(t)
		})
	}
}

type mockRun struct {
	commands     []string
	got          []string
	returnString []string
	returnErr    error
}

func (m *mockRun) fn() func(context.Context, ...string) (string, error) {
	return func(_ context.Context, commands ...string) (string, error) {
		m.got = commands
		return strings.Join(m.returnString, "\n"), m.returnErr
	}
}

func (m *mockRun) verify(t *testing.T) {
	assertStringsEqual(t, m.commands, m.got)
}

func assertStringsEqual(t *testing.T, expect, got []string) {
	expectString := strings.Join(expect, "\n")
	gotString := strings.Join(got, "\n")
	assert.Equal(t, expectString, gotString)
}
