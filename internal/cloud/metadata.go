package cloud

import (
	"os/exec"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	Azure = "azure"
	Aws   = "aws"
	Gcp   = "gcp"
	// DMI chassis asset tag for Azure machines, needed to identify wether or not we are running on Azure
	// This is actually ASCII-encoded, the decoding into a string results in "MSFT AZURE VM"
	azureDmiTag = "7783-7084-3265-9085-8269-3286-77"
)

type Instance struct {
	Provider string      `mapstructure:"provider,omitempty"`
	Metadata interface{} `mapstructure:"metadata,omitempty"`
}

type CustomCommand func(name string, arg ...string) *exec.Cmd

var customExecCommand CustomCommand = exec.Command //nolint

// All these detection methods are based in crmsh code, which has been refined over the years
// https://github.com/ClusterLabs/crmsh/blob/master/crmsh/utils.py#L2009

func identifyAzure() (bool, error) {
	log.Debug("Checking if the VM is running on Azure...")
	output, err := customExecCommand("dmidecode", "-s", "chassis-asset-tag").Output()
	if err != nil {
		return false, err
	}

	provider := strings.TrimSpace(string(output))
	log.Debugf("dmidecode output: %s", provider)

	return provider == azureDmiTag, nil
}

func identifyAws() (bool, error) {
	log.Debug("Checking if the VM is running on Aws...")
	systemVersion, err := customExecCommand("dmidecode", "-s", "system-version").Output()
	if err != nil {
		return false, err
	}

	systemVersionTrim := strings.ToLower(strings.TrimSpace(string(systemVersion)))
	log.Debugf("dmidecode system-version output: %s", systemVersionTrim)

	result, _ := regexp.MatchString(".*amazon.*", systemVersionTrim)
	if result {
		return result, nil
	}

	systemManufacturer, err := customExecCommand("dmidecode", "-s", "system-manufacturer").Output()
	if err != nil {
		return false, err
	}

	systemManufacturerTrim := strings.ToLower(strings.TrimSpace(string(systemManufacturer)))
	log.Debugf("dmidecode system-manufacturer output: %s", systemManufacturerTrim)

	result, _ = regexp.MatchString(".*amazon.*", systemManufacturerTrim)

	return result, nil
}

func identifyGcp() (bool, error) {
	log.Debug("Checking if the VM is running on Gcp...")
	output, err := customExecCommand("dmidecode", "-s", "bios-vendor").Output()
	if err != nil {
		return false, err
	}

	provider := strings.TrimSpace(string(output))
	log.Debugf("dmidecode output: %s", provider)

	return regexp.MatchString(".*Google.*", provider)
}

func IdentifyCloudProvider() (string, error) {
	log.Info("Identifying if the VM is running in a cloud environment...")

	if result, err := identifyAzure(); err != nil {
		return "", err
	} else if result {
		log.Infof("VM is running on %s", Azure)
		return Azure, nil
	}

	if result, err := identifyAws(); err != nil {
		return "", err
	} else if result {
		log.Infof("VM is running on %s", Aws)
		return Aws, nil
	}

	if result, err := identifyGcp(); err != nil {
		return "", err
	} else if result {
		log.Infof("VM is running on %s", Gcp)
		return Gcp, nil
	}

	log.Info("VM is not running in any recognized cloud provider")
	return "", nil
}

func NewCloudInstance() (*Instance, error) {
	var err error
	var cloudMetadata interface{}

	provider, err := IdentifyCloudProvider()
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
	case Aws:
		awsMetadata, err := NewAwsMetadata()
		if err != nil {
			return nil, err
		}
		cloudMetadata = NewAwsMetadataDto(awsMetadata)
	case Gcp:
		gcpMetadata, err := NewGcpMetadata()
		if err != nil {
			return nil, err
		}
		cloudMetadata = NewGcpMetadataDto(gcpMetadata)
	}

	cInst.Metadata = cloudMetadata

	return cInst, nil

}
