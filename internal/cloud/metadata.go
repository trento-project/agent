package cloud

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/utils"
)

const (
	Azure = "azure"
	AWS   = "aws"
	GCP   = "gcp"
	// DMI chassis asset tag for Azure machines, needed to identify wether or not we are running on Azure
	// This is actually ASCII-encoded, the decoding into a string results in "MSFT AZURE VM"
	azureDmiTag = "7783-7084-3265-9085-8269-3286-77"
)

type Instance struct {
	Provider string      `mapstructure:"provider,omitempty"`
	Metadata interface{} `mapstructure:"metadata,omitempty"`
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
	log.Debug("Checking if the VM is running on Azure...")
	output, err := i.executor.Exec("dmidecode", "-s", "chassis-asset-tag")
	if err != nil {
		return false, err
	}

	provider := strings.TrimSpace(string(output))
	log.Debugf("dmidecode output: %s", provider)

	return provider == azureDmiTag, nil
}

func (i *Identifier) identifyAWS() (bool, error) {
	log.Debug("Checking if the VM is running on Aws...")
	systemVersion, err := i.executor.Exec("dmidecode", "-s", "system-version")
	if err != nil {
		return false, err
	}

	systemVersionTrim := strings.ToLower(strings.TrimSpace(string(systemVersion)))
	log.Debugf("dmidecode system-version output: %s", systemVersionTrim)

	result, _ := regexp.MatchString(".*amazon.*", systemVersionTrim)
	if result {
		return result, nil
	}

	systemManufacturer, err := i.executor.Exec("dmidecode", "-s", "system-manufacturer")
	if err != nil {
		return false, err
	}

	systemManufacturerTrim := strings.ToLower(strings.TrimSpace(string(systemManufacturer)))
	log.Debugf("dmidecode system-manufacturer output: %s", systemManufacturerTrim)

	result, _ = regexp.MatchString(".*amazon.*", systemManufacturerTrim)

	return result, nil
}

func (i *Identifier) identifyGCP() (bool, error) {
	log.Debug("Checking if the VM is running on Gcp...")
	output, err := i.executor.Exec("dmidecode", "-s", "bios-vendor")
	if err != nil {
		return false, err
	}

	provider := strings.TrimSpace(string(output))
	log.Debugf("dmidecode output: %s", provider)

	return regexp.MatchString(".*Google.*", provider)
}

func (i *Identifier) IdentifyCloudProvider() (string, error) {
	log.Info("Identifying if the VM is running in a cloud environment...")

	if result, err := i.identifyAzure(); err != nil {
		return "", err
	} else if result {
		log.Infof("VM is running on %s", Azure)
		return Azure, nil
	}

	if result, err := i.identifyAWS(); err != nil {
		return "", err
	} else if result {
		log.Infof("VM is running on %s", AWS)
		return AWS, nil
	}

	if result, err := i.identifyGCP(); err != nil {
		return "", err
	} else if result {
		log.Infof("VM is running on %s", GCP)
		return GCP, nil
	}

	log.Info("VM is not running in any recognized cloud provider")
	return "", nil
}

func NewCloudInstance(commandExecutor utils.CommandExecutor) (*Instance, error) {
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
		cloudMetadata, err = NewAzureMetadata()
		if err != nil {
			return nil, err
		}
	case AWS:
		awsMetadata, err := NewAWSMetadata()
		if err != nil {
			return nil, err
		}
		cloudMetadata = NewAWSMetadataDto(awsMetadata)
	case GCP:
		gcpMetadata, err := NewGCPMetadata()
		if err != nil {
			return nil, err
		}
		cloudMetadata = NewGCPMetadataDto(gcpMetadata)
	}

	cInst.Metadata = cloudMetadata

	return cInst, nil

}
