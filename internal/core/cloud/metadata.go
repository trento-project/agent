package cloud

import (
	"context"
	"regexp"
	"strings"

	"log/slog"

	"github.com/trento-project/agent/pkg/utils"
)

const (
	Azure   = "azure"
	AWS     = "aws"
	GCP     = "gcp"
	Nutanix = "nutanix"
	KVM     = "kvm"
	VMware  = "vmware"

	// DMI chassis asset tag for Azure machines, needed to identify wether or not we are running on Azure
	// This is actually ASCII-encoded, the decoding into a string results in "MSFT AZURE VM"
	azureDmiTag = "7783-7084-3265-9085-8269-3286-77"
)

type Instance struct {
	Provider string
	Metadata interface{}
}

type Identifier struct {
	executor utils.CommandExecutor
}

func NewIdentifier(executor utils.CommandExecutor) *Identifier {
	return &Identifier{
		executor: executor,
	}
}

// All these detection methods are based in crmsh code, which has been refined over the years
// https://github.com/ClusterLabs/crmsh/blob/master/crmsh/utils.py#L2009

func (i *Identifier) identifyAzure() (bool, error) {
	slog.Debug("Checking if the system is running on Azure...")
	output, err := i.executor.Output("/usr/sbin/dmidecode", "-s", "chassis-asset-tag")
	if err != nil {
		return false, err
	}

	provider := strings.TrimSpace(string(output))
	slog.Debug("dmidecode result", "output", provider)

	return provider == azureDmiTag, nil
}

func (i *Identifier) identifyAWS() (bool, error) {
	slog.Debug("Checking if the system is running on Aws...")
	systemVersion, err := i.executor.Output("/usr/sbin/dmidecode", "-s", "system-version")
	if err != nil {
		return false, err
	}

	systemVersionTrim := strings.ToLower(strings.TrimSpace(string(systemVersion)))
	slog.Debug("dmidecode system-version", "output", systemVersionTrim)

	result, _ := regexp.MatchString(".*amazon.*", systemVersionTrim)
	if result {
		return result, nil
	}

	systemManufacturer, err := i.executor.Output("/usr/sbin/dmidecode", "-s", "system-manufacturer")
	if err != nil {
		return false, err
	}

	systemManufacturerTrim := strings.ToLower(strings.TrimSpace(string(systemManufacturer)))
	slog.Debug("dmidecode system-manufacturer", "output", systemManufacturerTrim)

	result, _ = regexp.MatchString(".*amazon.*", systemManufacturerTrim)

	return result, nil
}

func (i *Identifier) identifyGCP() (bool, error) {
	slog.Debug("Checking if the system is running on Gcp...")
	output, err := i.executor.Output("/usr/sbin/dmidecode", "-s", "bios-vendor")
	if err != nil {
		return false, err
	}

	provider := strings.TrimSpace(string(output))
	slog.Debug("dmidecode", "output", provider)

	return regexp.MatchString(".*Google.*", provider)
}

func (i *Identifier) identifyNutanix() (bool, error) {
	slog.Debug("Checking if the system is running on Nutanix...")
	output, err := i.executor.Output("/usr/sbin/dmidecode")
	if err != nil {
		return false, err
	}

	dmidecodeContent := strings.TrimSpace(string(output))
	slog.Debug("dmidecode", "output", dmidecodeContent)

	return regexp.MatchString("(?i)nutanix|ahv", dmidecodeContent)
}

func (i *Identifier) identifyKVM() (bool, error) {
	slog.Debug("Checking if the system is running under KVM...")
	output, err := i.executor.Output("/usr/bin/systemd-detect-virt")
	if err != nil {
		return false, err
	}

	systemdDetectVirtContent := strings.TrimSpace(string(output))
	slog.Debug("systemd-detect-virt", "output", systemdDetectVirtContent)

	return systemdDetectVirtContent == KVM, nil
}

func (i *Identifier) identifyVMware() (bool, error) {
	slog.Debug("Checking if the system is running under VMware...")
	output, err := i.executor.Output("/usr/bin/systemd-detect-virt")
	if err != nil {
		return false, err
	}

	systemdDetectVirtContent := strings.TrimSpace(string(output))
	slog.Debug("systemd-detect-virt", "output", systemdDetectVirtContent)

	return systemdDetectVirtContent == VMware, nil
}

func (i *Identifier) IdentifyCloudProvider() (string, error) {
	slog.Info("Identifying if the system is running in a cloud environment...")

	providers := []struct {
		identifyFn func() (bool, error)
		name       string
	}{
		{identifyFn: i.identifyAzure, name: Azure},
		{identifyFn: i.identifyAWS, name: AWS},
		{identifyFn: i.identifyGCP, name: GCP},
		{identifyFn: i.identifyNutanix, name: Nutanix},
		{identifyFn: i.identifyKVM, name: KVM},
		{identifyFn: i.identifyVMware, name: VMware},
	}

	for _, provider := range providers {
		if result, err := provider.identifyFn(); err != nil {
			return "", err
		} else if result {
			slog.Info("System is running on a known provider", "provider", provider.name)
			return provider.name, nil
		}
	}
	slog.Info("The system is not running in any recognized cloud provider")
	return "", nil
}

func NewCloudInstance(
	ctx context.Context,
	commandExecutor utils.CommandExecutor,
	client HTTPClient,
) (*Instance, error) {
	var err error
	var cloudMetadata interface{}

	cloudIdentifier := NewIdentifier(commandExecutor)

	provider, err := cloudIdentifier.IdentifyCloudProvider()
	if err != nil {
		return nil, err
	}

	cInst := &Instance{
		Metadata: nil,
		Provider: provider,
	}

	switch provider {
	case Azure:
		{
			cloudMetadata, err = NewAzureMetadata(ctx, client)
			if err != nil {
				return nil, err
			}
		}
	case AWS:
		{
			awsMetadata, err := NewAWSMetadata(ctx, client)
			if err == nil {
				cloudMetadata = NewAWSMetadataDto(awsMetadata)
			}
		}
	case GCP:
		{
			gcpMetadata, err := NewGCPMetadata(ctx, client)
			if err != nil {
				return nil, err
			}
			cloudMetadata = NewGCPMetadataDto(gcpMetadata)
		}
	}

	cInst.Metadata = cloudMetadata

	return cInst, nil

}
