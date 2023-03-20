package version

import (
	"context"
	"flag"
	"fmt"

	"gitlab.com/gitlab-org/fleeting/nesting"
)

type versionCmd struct {
	fs *flag.FlagSet
}

func New() *versionCmd {
	c := &versionCmd{}
	c.fs = flag.NewFlagSet("version", flag.ExitOnError)

	return c
}

func (cmd *versionCmd) Command() (*flag.FlagSet, string) {
	return cmd.fs, ""
}

func (cmd *versionCmd) Execute(ctx context.Context) error {
	fmt.Println(nesting.Version.Full())

	return nil
}
