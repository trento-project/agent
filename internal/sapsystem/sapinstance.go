package sapsystem

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/trento-project/agent/internal/sapsystem/sapcontrolapi"
	"github.com/trento-project/agent/internal/utils"
)

var (
	databaseFeatures         = regexp.MustCompile("HDB.*")
	applicationFeatures      = regexp.MustCompile("MESSAGESERVER.*|ENQREP|ABAP.*")
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

func NewSAPInstance(w sapcontrolapi.WebService) (*SAPInstance, error) {
	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	scontrol, err := NewSAPControl(w)
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
		sid, _ := sapInstance.SAPControl.findProperty("SAPSYSTEMNAME")
		sapInstance.SystemReplication = systemReplicationStatus(sid, sapInstance.Name)
		sapInstance.HostConfiguration = landscapeHostConfiguration(sid, sapInstance.Name)
		sapInstance.HdbnsutilSRstate = hdbnsutilSrstate(sid, sapInstance.Name)
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

func runPythonSupport(sid, instance, script string) map[string]interface{} {
	user := fmt.Sprintf("%sadm", strings.ToLower(sid))
	cmdPath := path.Join(sapInstallationPath, sid, instance, "exe/python_support", script)
	cmd := fmt.Sprintf("python %s --sapcontrol=1", cmdPath)
	// Even with a error return code, some data is available
	srData, _ := customExecCommand("su", "-lc", cmd, user).Output()

	dataMap := utils.FindMatches(`(\S+)=(.*)`, srData)

	return dataMap
}

func systemReplicationStatus(sid, instance string) map[string]interface{} {
	return runPythonSupport(sid, instance, "systemReplicationStatus.py")
}

func landscapeHostConfiguration(sid, instance string) map[string]interface{} {
	return runPythonSupport(sid, instance, "landscapeHostConfiguration.py")
}

func hdbnsutilSrstate(sid, instance string) map[string]interface{} {
	user := fmt.Sprintf("%sadm", strings.ToLower(sid))
	cmdPath := path.Join(sapInstallationPath, sid, instance, "exe", "hdbnsutil")
	cmd := fmt.Sprintf("%s -sr_state -sapcontrol=1", cmdPath)
	srData, _ := customExecCommand("su", "-lc", cmd, user).Output()
	dataMap := utils.FindMatches(`(.+)=(.*)`, srData)
	return dataMap
}
