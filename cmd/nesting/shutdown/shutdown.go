package shutdown

import (
	"context"
	"flag"

	"gitlab.com/ajwalker/nesting/api"
)

type shutdownCmd struct {
	fs *flag.FlagSet

	configPath string
}

func New() *shutdownCmd {
	c := &shutdownCmd{}
	c.fs = flag.NewFlagSet("shutdown", flag.ExitOnError)

	return c
}

func (cmd *shutdownCmd) Command() (*flag.FlagSet, string) {
	return cmd.fs, ""
}

func (cmd *shutdownCmd) Execute(ctx context.Context) error {
	conn, err := api.DefaultConn()
	if err != nil {
		return err
	}

	err = api.New(conn).Shutdown(ctx)
	if err != nil {
		return err
	}

	return nil
}
