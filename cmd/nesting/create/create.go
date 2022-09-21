package create

import (
	"context"
	"flag"
	"fmt"

	"gitlab.com/ajwalker/nesting/api"
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
	return cmd.fs, "<image name>"
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

	vm, err := client.Create(ctx, cmd.fs.Args()[0])
	if err != nil {
		return err
	}

	fmt.Println(vm.GetId(), vm.GetName(), vm.GetAddr())

	return nil
}
