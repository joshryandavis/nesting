package delete

import (
	"context"
	"flag"

	"gitlab.com/gitlab-org/fleeting/nesting/api"
)

type deleteCmd struct {
	fs *flag.FlagSet
}

func New() *deleteCmd {
	c := &deleteCmd{}
	c.fs = flag.NewFlagSet("delete", flag.ExitOnError)
	return c
}

func (cmd *deleteCmd) Command() (*flag.FlagSet, string) {
	return cmd.fs, "<image id>"
}

func (cmd *deleteCmd) Execute(ctx context.Context) error {
	if len(cmd.fs.Args()) < 1 {
		return flag.ErrHelp
	}

	conn, err := api.DefaultConn()
	if err != nil {
		return err
	}

	client := api.New(conn)
	defer client.Close()

	return client.Delete(ctx, cmd.fs.Args()[0])
}
