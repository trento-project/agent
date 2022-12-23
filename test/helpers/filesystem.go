package helpers

import (
	"github.com/spf13/afero"
)

const (
	DummyMachineID = "dummy-machine-id"
	DummyAgentID   = "779cdd70-e9e2-58ca-b18a-bf3eb3f71244"

	machineIDPath = "/etc/machine-id"
)

// MockMachineIDFile mocks the /etc/machine-id file to have a known value
func MockMachineIDFile() afero.Fs {
	fileSystem := afero.NewMemMapFs()

	err := afero.WriteFile(fileSystem, machineIDPath, []byte(DummyMachineID), 0644)

	if err != nil {
		panic(err)
	}

	return fileSystem
}
