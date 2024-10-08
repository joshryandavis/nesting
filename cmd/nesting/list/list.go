package list

import (
	"context"
	"flag"
	"fmt"

	"gitlab.com/gitlab-org/fleeting/nesting/api"
)

type listCmd struct {
	fs *flag.FlagSet
}

func New() *listCmd {
	c := &listCmd{}
	c.fs = flag.NewFlagSet("list", flag.ExitOnError)
	return c
}

func (cmd *listCmd) Command() (*flag.FlagSet, string) {
	return cmd.fs, ""
}

func (cmd *listCmd) Execute(ctx context.Context) error {
	conn, err := api.DefaultConn()
	if err != nil {
		return err
	}

	client := api.New(conn)
	defer client.Close()

	vms, err := client.List(ctx)
	if err != nil {
		return err
	}

	for _, vm := range vms {
		fmt.Println(vm.GetId(), vm.GetName(), vm.GetAddr())
	}

	return nil
}
