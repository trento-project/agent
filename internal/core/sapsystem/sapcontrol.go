package sapsystem

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/spf13/afero"
	sapcontrol "github.com/trento-project/agent/internal/core/sapsystem/sapcontrolapi"
)

type SAPControl struct {
	Processes  []*sapcontrol.OSProcess
	Instances  []*sapcontrol.SAPInstance
	Properties []*sapcontrol.InstanceProperty
}

func NewSAPControl(ctx context.Context, w sapcontrol.WebService, fs afero.Fs, hostname string) (*SAPControl, error) {
	properties, err := w.GetInstanceProperties(ctx)
	if err != nil {
		return nil, fmt.Errorf("SAPControl web service error: %w", err)
	}

	processes, err := w.GetProcessList(ctx)
	if err != nil {
		return nil, fmt.Errorf("SAPControl web service error: %w", err)
	}

	instances, err := w.GetSystemInstanceList(ctx)
	if err != nil {
		return nil, fmt.Errorf("SAPControl web service error: %w", err)
	}

	sapControl := &SAPControl{
		Properties: properties.Properties,
		Processes:  processes.Processes,
		Instances:  instances.Instances,
	}

	if err := sapControl.enrichCurrentInstance(fs, hostname); err != nil {
		return nil, fmt.Errorf("Error finding current instance: %w", err)
	}

	return sapControl, nil
}

func (s *SAPControl) findProperty(key string) (string, error) {
	for _, item := range s.Properties {
		if item.Property == key {
			return item.Value, nil
		}
	}

	return "", fmt.Errorf("Property %s not found", key)
}

// enrichCurrentInstance identifies and sets the currently discovered instance in the
// SapControl.Instances list. This is required to later on extract information from
// that dataset
// The logic is based on this: https://m1bc.home.blog/2019/09/09/getsysteminstancelist-duplicate-entries/
// The file syntax is: startPriority_httpPort_httpsPort_features_dispstatus_instanceNr_hostname
// The content of the file includes the real hostname of the machine where the instance is running
func (s *SAPControl) enrichCurrentInstance(fs afero.Fs, hostname string) error {
	sid, err := s.findProperty("SAPSYSTEMNAME")
	if err != nil {
		return err
	}

	instanceNumber, err := s.findProperty("SAPSYSTEM")
	if err != nil {
		return err
	}

	sapLocalhost, err := s.findProperty("SAPLOCALHOST")
	if err != nil {
		return err
	}

	sapControlInstancesPath := path.Join("/usr/sap", sid, "/SYS/global/sapcontrol")
	instanceFiles, err := afero.ReadDir(fs, sapControlInstancesPath)
	if err != nil {
		return fmt.Errorf("sapcontrol folder not found: %w", err)
	}

	for _, instance := range s.Instances {
		if fmt.Sprintf("%02d", instance.InstanceNr) != instanceNumber || instance.Hostname != sapLocalhost {
			continue
		}

		for _, instanceFile := range instanceFiles {
			if !instanceMatches(instance, instanceFile.Name()) {
				continue
			}
			absInstancePath := path.Join(sapControlInstancesPath, instanceFile.Name())
			instanceFileContent, err := afero.ReadFile(fs, absInstancePath)
			if err != nil {
				continue
			}

			if strings.Contains(string(instanceFileContent), hostname) {
				instance.CurrentInstance = true
				break
			}
		}
	}

	return nil
}

func instanceMatches(instance *sapcontrol.SAPInstance, instanceFile string) bool {
	matched, err := regexp.MatchString(fmt.Sprintf(
		"%s_%d_%d_.*_%d_%02d_%s",
		instance.StartPriority,
		instance.HttpPort,
		instance.HttpsPort,
		sapcontrol.DispstatusCodeFromStr(instance.Dispstatus),
		instance.InstanceNr,
		instance.Hostname,
	), instanceFile)

	if err != nil {
		return false
	}

	return matched
}
