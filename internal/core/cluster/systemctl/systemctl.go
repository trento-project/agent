package systemctl

import (
	"github.com/trento-project/agent/pkg/utils"
)

type Systemctl struct {
	CommandExecutor utils.CommandExecutor
}

func NewSystemctl(commandExecutor utils.CommandExecutor) *Systemctl {
	return &Systemctl{
		CommandExecutor: commandExecutor,
	}
}

func (s *Systemctl) IsActive(serviceName string) bool {
	_, err := s.CommandExecutor.Output("systemctl", "is-active", serviceName)
	// If the service is active, the command returns 0 exit code.
	return err == nil
}
