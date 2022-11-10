# nesting

`nesting` is a basic and opinionated daemon that sits in front of virtualization
platforms.

Its goal is to provide a client/server model for listing, creating and deleting
pre-configured Virtual Machines intended for isolated and short-lived workloads.

## Platforms supported

### MacOS Host

- Tart (m1)
- Parallels (intel)

## Usage

### CLI

```shell
$ ./nesting --help
serve
  -config string
        config
  -hypervisor string
        hypervisor (default "parallels")
init
  -config string
        config
shutdown
create <image name>
delete <image id>
list 
```

### Client example

```golang
import (
	"fmt"
	"context"

	"gitlab.com/gitlab-org/fleeting/nesting/api"
)

func Example(ctx context.Context) {
	conn, err := api.DefaultConn()
	if err != nil {
		return err
	}

	cli := api.New(conn)

	// cli.Init(ctx)
	// vm, err := cli.Create(ctx, "image")
	// defer cli.Delete(vm.GetId())

	vms, err := cli.List(ctx)
	if err != nil {
		return err
	}

	for _, vm := range vms {
		fmt.Println(vm.GetId(), vm.GetName(), vm.GetAddr())
	}
}
```
