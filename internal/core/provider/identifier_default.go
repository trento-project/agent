//go:build !ppc64le

package provider

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/utils"
)

// DMI chassis asset tag for Azure machines, needed to identify wether or not we are running on Azure
// This is actually ASCII-encoded, the decoding into a string results in "MSFT AZURE VM"
const azureDmiTag = "7783-7084-3265-9085-8269-3286-77"

func NewIdentifier(executor utils.CommandExecutor) Identifier {
	return &identifier{
		executor: executor,
	}
}

func (i *identifier) IdentifyProvider() (string, error) {
	log.Info("NOT ON PPC")
	log.Info("Identifying if the system is running in a cloud environment...")

	if result, err := i.identifyAzure(); err != nil {
		return "", err
	} else if result {
		log.Infof("System is running on %s", Azure)
		return Azure, nil
	}

	if result, err := i.identifyAWS(); err != nil {
		return "", err
	} else if result {
		log.Infof("System is running on %s", AWS)
		return AWS, nil
	}

	if result, err := i.identifyGCP(); err != nil {
		return "", err
	} else if result {
		log.Infof("System is running on %s", GCP)
		return GCP, nil
	}

	if result, err := i.identifyNutanix(); err != nil {
		return "", err
	} else if result {
		log.Infof("System is running on %s", Nutanix)
		return Nutanix, nil
	}

	if result, err := i.identifyKVM(); err != nil {
		return "", err
	} else if result {
		log.Infof("System is running on %s", KVM)
		return KVM, nil
	}

	if result, err := i.identifyVMware(); err != nil {
		return "", err
	} else if result {
		log.Infof("System is running on %s", VMware)
		return VMware, nil
	}

	log.Info("The system is not running in any recognized cloud provider")
	return "", nil
}

// All these detection methods are based in crmsh code, which has been refined over the years
// https://github.com/ClusterLabs/crmsh/blob/master/crmsh/utils.py#L2009

func (i *identifier) identifyAzure() (bool, error) {
	log.Debug("Checking if the system is running on Azure...")
	output, err := i.executor.Exec("dmidecode", "-s", "chassis-asset-tag")
	if err != nil {
		return false, err
	}

	provider := strings.TrimSpace(string(output))
	log.Debugf("dmidecode output: %s", provider)

	return provider == azureDmiTag, nil
}

func (i *identifier) identifyAWS() (bool, error) {
	log.Debug("Checking if the system is running on Aws...")
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

func (i *identifier) identifyGCP() (bool, error) {
	log.Debug("Checking if the system is running on Gcp...")
	output, err := i.executor.Exec("dmidecode", "-s", "bios-vendor")
	if err != nil {
		return false, err
	}

	provider := strings.TrimSpace(string(output))
	log.Debugf("dmidecode output: %s", provider)

	return regexp.MatchString(".*Google.*", provider)
}

func (i *identifier) identifyNutanix() (bool, error) {
	log.Debug("Checking if the system is running on Nutanix...")
	output, err := i.executor.Exec("dmidecode")
	if err != nil {
		return false, err
	}

	dmidecodeContent := strings.TrimSpace(string(output))
	log.Debugf("dmidecode output: %s", dmidecodeContent)

	return regexp.MatchString("(?i)nutanix|ahv", dmidecodeContent)
}

func (i *identifier) identifyKVM() (bool, error) {
	log.Debug("Checking if the system is running under KVM...")
	output, err := i.executor.Exec("systemd-detect-virt")
	if err != nil {
		return false, err
	}

	systemdDetectVirtContent := strings.TrimSpace(string(output))
	log.Debugf("systemd-detect-virt output: %s", systemdDetectVirtContent)

	return systemdDetectVirtContent == KVM, nil
}

func (i *identifier) identifyVMware() (bool, error) {
	log.Debug("Checking if the system is running under VMware...")
	output, err := i.executor.Exec("systemd-detect-virt")
	if err != nil {
		return false, err
	}

	systemdDetectVirtContent := strings.TrimSpace(string(output))
	log.Debugf("systemd-detect-virt output: %s", systemdDetectVirtContent)

	return systemdDetectVirtContent == VMware, nil
}
