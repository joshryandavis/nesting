package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"gitlab.com/gitlab-org/fleeting/nesting/cmd/nesting/create"
	"gitlab.com/gitlab-org/fleeting/nesting/cmd/nesting/delete"
	"gitlab.com/gitlab-org/fleeting/nesting/cmd/nesting/initialize"
	"gitlab.com/gitlab-org/fleeting/nesting/cmd/nesting/list"
	"gitlab.com/gitlab-org/fleeting/nesting/cmd/nesting/serve"
	"gitlab.com/gitlab-org/fleeting/nesting/cmd/nesting/shutdown"
	"gitlab.com/gitlab-org/fleeting/nesting/cmd/nesting/version"
)

type Command interface {
	Command() (fs *flag.FlagSet, usage string)
	Execute(ctx context.Context) error
}

type Commands []Command

func (c Commands) Usage() {
	for _, cmd := range c {
		fs, usage := cmd.Command()
		fmt.Fprintln(fs.Output(), fs.Name(), usage)
		fs.PrintDefaults()
	}
	os.Exit(1)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cmds := Commands{
		serve.New(),
		initialize.New(),
		shutdown.New(),
		create.New(),
		delete.New(),
		list.New(),
		version.New(),
	}

	if len(os.Args) < 2 {
		cmds.Usage()
	}

	for _, cmd := range cmds {
		fs, usage := cmd.Command()
		if os.Args[1] == fs.Name() {
			fs.Parse(os.Args[2:])
			if err := cmd.Execute(ctx); err != nil {
				if errors.Is(err, flag.ErrHelp) {
					fmt.Fprintln(fs.Output(), fs.Name(), usage)
					fs.PrintDefaults()
					os.Exit(1)
				}

				panic(err)
			}
			return
		}
	}

	cmds.Usage()
}
