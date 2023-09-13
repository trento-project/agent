package saptune

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/pkg/utils/mocks"
)

type SaptuneTestSuite struct {
	suite.Suite
}

func TestSaptuneTestSuite(t *testing.T) {
	suite.Run(t, new(SaptuneTestSuite))
}

func (suite *SaptuneTestSuite) TestNewSaptune() {
	mockCommand := new(mocks.CommandExecutor)

	mockCommand.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		[]byte("3.1.0"), nil,
	)

	saptuneOutput := []byte(`{"$schema":"file:///usr/share/saptune/schemas/1.0/saptune_status.schema.json","publish time":"2023-09-12 16:28:11.718","argv":"saptune --format json status","pid":6675,"command":"status","exit code":1,"result":{"services":{"saptune":["disabled","inactive"],"sapconf":[],"tuned":[]},"systemd system state":"degraded","tuning state":"compliant","virtualization":"kvm","configured version":"3","package version":"3.1.0","Solution enabled":[],"Notes enabled by Solution":[],"Solution applied":[],"Notes applied by Solution":[],"Notes enabled additionally":["1410736"],"Notes enabled":["1410736"],"Notes applied":["1410736"],"staging":{"staging enabled":false,"Notes staged":[],"Solutions staged":[]},"remember message":"\nRemember: if you wish to automatically activate the note's and solution's tuning options after a reboot, you must enable and start saptune.service by running:\n 'saptune service enablestart'.\nThe systemd system state is NOT ok.\nPlease call 'saptune check' to get guidance to resolve the issues!\n\n"},"messages":[{"priority":"NOTICE","message":"actions.go:85: ATTENTION: You are running a test version (3.1.0 from 2023/08/03) of saptune which is not supported for production use\n"}]}`)

	mockCommand.On("Exec", "saptune", "--format", "json", "status").Return(
		saptuneOutput, nil,
	)

	expectedVersion := "3.1.0"
	saptuneDetails, err := NewSaptune(mockCommand)

	expectedDetails := Saptune{
		Version: &expectedVersion,
		Commands: []SaptuneOutput{
			{
				Schema:      "file:///usr/share/saptune/schemas/1.0/saptune_status.schema.json",
				PublishTime: "2023-09-12 16:28:11.718",
				Argv:        "saptune --format json status",
				Pid:         6675,
				Command:     "status",
				ExitCode:    1,
				Result: Result{
					Services: Services{
						Saptune: []string{"disabled", "inactive"},
						Sapconf: []string{},
						Tuned:   []string{},
					},
					SystemdSystemState: "degraded",
					TuningState: "compliant",
					Virtualization: "kvm",
					ConfiguredVersion: "3",
					PackageVersion: "3.1.0",
					SolutionEnabled: []string{},
					NotesEnabledBySolution: []string{},
					SolutionApplied: []string{},
					NotesAppliedBySolution: []string{},
					NotesEnabledAdditionally: []string{"1410736"},
					NotesEnabled: []string{"1410736"},
					NotesApplied: []string{"1410736"},
					Staging: Staging{
						StagingEnabled: false,
						NotesStaged: []string{},
						SolutionsStaged: []string{},
					},
					RememberMessage:"\nRemember: if you wish to automatically activate the note's and solution's tuning options after a reboot, you must enable and start saptune.service by running:\n 'saptune service enablestart'.\nThe systemd system state is NOT ok.\nPlease call 'saptune check' to get guidance to resolve the issues!\n\n",
				},
				Messages: []Message{
					{
						Priority: "NOTICE",
						Message: "actions.go:85: ATTENTION: You are running a test version (3.1.0 from 2023/08/03) of saptune which is not supported for production use\n",
					},
				},
			},
		},
	}

	suite.NoError(err)
	suite.Equal(expectedDetails, saptuneDetails)
}

func (suite *SaptuneTestSuite) TestNewSaptuneSaptuneVersionUnknownErr() {
	mockCommand := new(mocks.CommandExecutor)

	mockCommand.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		nil, errors.New("Error: exec: \"rpm\": executable file not found in $PATH"),
	)
	
	saptuneDetails, err := NewSaptune(mockCommand)

	expectedDetails := Saptune{
		Version: nil,
		Commands: []SaptuneOutput{},
	}
	suite.EqualError(err, ErrSaptuneVersionUnknown.Error())
	suite.Equal(expectedDetails, saptuneDetails)
}

func (suite *SaptuneTestSuite) TestNewSaptuneUnsupportedSaptuneVerErr() {
	mockCommand := new(mocks.CommandExecutor)

	mockCommand.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "saptune").Return(
		[]byte("3.0.0"), nil,
	)

	saptuneDetails, err := NewSaptune(mockCommand)

	expectedVersion := "3.0.0"
	expectedDetails := Saptune{
		Version: &expectedVersion,
		Commands: []SaptuneOutput{},
	}
	suite.EqualError(err, ErrUnsupportedSaptuneVer.Error())
	suite.Equal(expectedDetails, saptuneDetails)
}