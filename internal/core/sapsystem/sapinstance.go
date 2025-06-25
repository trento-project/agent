package sapsystem

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/trento-project/agent/internal/core/sapsystem/sapcontrolapi"
	"github.com/trento-project/agent/pkg/utils"
)

var (
	databaseFeatures         = regexp.MustCompile("HDB.*")
	applicationFeatures      = regexp.MustCompile("MESSAGESERVER.*|ENQREP|ABAP.*|J2EE.*")
	diagnosticsAgentFeatures = regexp.MustCompile("SMDAGENT")
)

type SystemReplication map[string]interface{}
type HostConfiguration map[string]interface{}
type HdbnsutilSRstate map[string]interface{}

type SAPInstance struct {
	Name       string
	Type       SystemType
	Host       string
	SAPControl *SAPControl
	// Only for Database type
	SystemReplication SystemReplication
	HostConfiguration HostConfiguration
	HdbnsutilSRstate  HdbnsutilSRstate
}

func NewSAPInstance(
	ctx context.Context,
	w sapcontrolapi.WebService,
	executor utils.CommandExecutor,
	fs afero.Fs,
) (*SAPInstance, error) {
	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	scontrol, err := NewSAPControl(ctx, w, fs, host)
	if err != nil {
		return nil, err
	}

	instanceName, err := scontrol.findProperty("INSTANCE_NAME")
	if err != nil {
		return nil, err
	}

	instanceType, err := detectType(scontrol)
	if err != nil {
		return nil, err
	}

	sapInstance := &SAPInstance{
		Host:              host,
		SAPControl:        scontrol,
		Name:              instanceName,
		Type:              instanceType,
		SystemReplication: nil,
		HostConfiguration: nil,
		HdbnsutilSRstate:  nil,
	}

	if instanceType == Database {
		sid, err := sapInstance.SAPControl.findProperty("SAPSYSTEMNAME")
		if err != nil {
			return nil, errors.Wrap(err, "Error finding the SAP instance sid")
		}

		sapInstance.SystemReplication = systemReplicationStatus(executor, sid, sapInstance.Name)
		sapInstance.HostConfiguration = landscapeHostConfiguration(executor, sid, sapInstance.Name)
		sapInstance.HdbnsutilSRstate = hdbnsutilSrstate(executor, sid, sapInstance.Name)
	}

	return sapInstance, nil
}

func detectType(sapControl *SAPControl) (SystemType, error) {
	sapLocalhost, err := sapControl.findProperty("SAPLOCALHOST")
	if err != nil {
		return Unknown, err
	}

	for _, instance := range sapControl.Instances {
		if instance.Hostname == sapLocalhost {
			switch {
			case databaseFeatures.MatchString(instance.Features):
				return Database, nil
			case applicationFeatures.MatchString(instance.Features):
				return Application, nil
			case diagnosticsAgentFeatures.MatchString(instance.Features):
				return DiagnosticsAgent, nil
			default:
				return Unknown, nil
			}
		}
	}

	return Unknown, nil
}

func runPythonSupport(executor utils.CommandExecutor, sid, instance, script string) map[string]interface{} {
	user := fmt.Sprintf("%sadm", strings.ToLower(sid))
	cmdPath := path.Join(sapInstallationPath, sid, instance, "exe/python_support", script)
	cmd := fmt.Sprintf("python %s --sapcontrol=1", cmdPath)
	// Even with a error return code, some data is available
	srData, err := executor.Exec("/usr/bin/su", "-lc", cmd, user)
	if err != nil {
		slog.Warn("Error running python_support command", "error", err)
	}
	dataMap := utils.FindMatches(`(\S+)=(.*)`, srData)

	return dataMap
}

func systemReplicationStatus(executor utils.CommandExecutor, sid, instance string) map[string]interface{} {
	return runPythonSupport(executor, sid, instance, "systemReplicationStatus.py")
}

func landscapeHostConfiguration(executor utils.CommandExecutor, sid, instance string) map[string]interface{} {
	return runPythonSupport(executor, sid, instance, "landscapeHostConfiguration.py")
}

func hdbnsutilSrstate(executor utils.CommandExecutor, sid, instance string) map[string]interface{} {
	user := fmt.Sprintf("%sadm", strings.ToLower(sid))
	cmdPath := path.Join(sapInstallationPath, sid, instance, "exe", "hdbnsutil")
	cmd := fmt.Sprintf("%s -sr_state -sapcontrol=1", cmdPath)
	srData, err := executor.Exec("/usr/bin/su", "-lc", cmd, user)
	if err != nil {
		slog.Warn("Error running hdbnsutil command", "error", err)
	}
	dataMap := utils.FindMatches(`(.+)=(.*)`, srData)
	return dataMap
}
