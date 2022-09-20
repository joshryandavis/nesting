package initialize

import (
	"context"
	"flag"
	"fmt"
	"os"

	"gitlab.com/ajwalker/nesting/api"
)

type initializeCmd struct {
	fs *flag.FlagSet

	configPath string
}

func New() *initializeCmd {
	c := &initializeCmd{}
	c.fs = flag.NewFlagSet("init", flag.ExitOnError)

	c.fs.StringVar(&c.configPath, "config", "", "config")

	return c
}

func (cmd *initializeCmd) Command() (*flag.FlagSet, string) {
	return cmd.fs, ""
}

func (cmd *initializeCmd) Execute(ctx context.Context) error {
	var (
		config []byte
		err    error
	)

	if cmd.configPath != "" {
		config, err = os.ReadFile(cmd.configPath)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
	}

	conn, err := api.DefaultConn()
	if err != nil {
		return err
	}

	err = api.New(conn).Init(ctx, config)
	if err != nil {
		return err
	}

	return nil
}
