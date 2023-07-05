//go:build !darwin

package virtualizationframework

import (
	"fmt"

	"gitlab.com/gitlab-org/fleeting/nesting/hypervisor"
)

func New(config []byte) (hypervisor.Hypervisor, error) {
	return nil, fmt.Errorf("unsupported hypervisor on this platform")
}
