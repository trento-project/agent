//go:build ppc64le

package provider

import (
	log "github.com/sirupsen/logrus"

	"github.com/trento-project/agent/pkg/utils"
)

func NewIdentifier(executor utils.CommandExecutor) Identifier {
	return &identifier{
		executor: executor,
	}
}

func (i *identifier) IdentifyProvider() (string, error) {
	log.Info("ON PPC")
	return Nutanix, nil
}
