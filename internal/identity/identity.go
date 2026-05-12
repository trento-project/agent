// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/afero"
)

const MachineIDPath = "/etc/machine-id"

var trentoNamespace = uuid.Must(uuid.Parse("fb92284e-aa5e-47f6-a883-bf9469e7a0dc")) //nolint:gochecknoglobals

func GetAgentID(fileSystem afero.Fs) (string, error) {
	machineIDBytes, err := afero.ReadFile(fileSystem, MachineIDPath)
	if err != nil {
		return "", err
	}

	machineID := strings.TrimSpace(string(machineIDBytes))
	agentID := uuid.NewSHA1(trentoNamespace, []byte(machineID))

	return agentID.String(), nil
}
