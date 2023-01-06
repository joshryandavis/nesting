package create

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	"gitlab.com/gitlab-org/fleeting/nesting/api"
)

type createCmd struct {
	fs *flag.FlagSet
}

func New() *createCmd {
	c := &createCmd{}
	c.fs = flag.NewFlagSet("create", flag.ExitOnError)
	return c
}

func (cmd *createCmd) Command() (*flag.FlagSet, string) {
	return cmd.fs, "<image name> [<slot number>]"
}

func (cmd *createCmd) Execute(ctx context.Context) error {
	if len(cmd.fs.Args()) < 1 {
		return flag.ErrHelp
	}

	conn, err := api.DefaultConn()
	if err != nil {
		return err
	}

	client := api.New(conn)
	defer client.Close()

	var slot *int32
	if len(cmd.fs.Args()) >= 2 {
		i, err := strconv.Atoi(cmd.fs.Args()[1])
		if err != nil {
			return err
		}
		s := int32(i)
		slot = &s
	}

	vm, stompedVmId, err := client.Create(ctx, cmd.fs.Args()[0], slot)
	if err != nil {
		return err
	}

	fmt.Println(vm.GetId(), vm.GetName(), vm.GetAddr())
	if stompedVmId != nil {
		fmt.Printf("stomped vm id %q\n", *stompedVmId)
	}

	return nil
}
