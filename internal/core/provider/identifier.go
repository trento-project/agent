package provider

import (
	"github.com/trento-project/agent/pkg/utils"
)

const (
	Azure   = "azure"
	AWS     = "aws"
	GCP     = "gcp"
	Nutanix = "nutanix"
	KVM     = "kvm"
	VMware  = "vmware"
)

type identifier struct {
	executor utils.CommandExecutor
}

type Identifier interface {
	IdentifyProvider() (string, error)
}
