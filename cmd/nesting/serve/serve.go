package serve

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	"gitlab.com/ajwalker/nesting/api"
	"gitlab.com/ajwalker/nesting/hypervisor"
	"gitlab.com/ajwalker/nesting/hypervisor/parallels"
	"gitlab.com/ajwalker/nesting/hypervisor/tart"
)

type serveCmd struct {
	fs         *flag.FlagSet
	hypervisor string
	configPath string
}

func New() *serveCmd {
	c := &serveCmd{}
	c.fs = flag.NewFlagSet("serve", flag.ExitOnError)

	switch runtime.GOOS {
	case "darwin":
		fallthrough
	default:
		c.hypervisor = "parallels"
	}

	c.fs.StringVar(&c.hypervisor, "hypervisor", c.hypervisor, "hypervisor")
	c.fs.StringVar(&c.configPath, "config", "", "config")

	return c
}

func (cmd *serveCmd) Command() (*flag.FlagSet, string) {
	return cmd.fs, ""
}

func (cmd *serveCmd) Execute(ctx context.Context) error {
	var (
		config []byte
		err    error
		hv     hypervisor.Hypervisor
	)

	if cmd.configPath != "" {
		config, err = os.ReadFile(cmd.configPath)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
	}

	switch cmd.hypervisor {
	case "parallels":
		hv, err = parallels.New(config)
		if err != nil {
			return err
		}

	case "tart":
		hv, err = tart.New(config)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown hypervisor %q", cmd.hypervisor)
	}

	return api.Serve(ctx, hv)
}
