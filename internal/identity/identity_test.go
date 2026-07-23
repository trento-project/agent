// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package identity_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/identity"
	"github.com/trento-project/agent/test/helpers"
)

type IdentityTestSuite struct {
	suite.Suite
}

func TestIdentityTestSuite(t *testing.T) {
	suite.Run(t, new(IdentityTestSuite))
}

func (suite *IdentityTestSuite) TestGetAgentID() {
	fileSystem := helpers.MockMachineIDFile()
	agentID, err := identity.GetAgentID(fileSystem)

	suite.Require().NoError(err)
	suite.Equal(helpers.DummyAgentID, agentID)
}

func (suite *IdentityTestSuite) TestGetAgentIDMachineIDNotFound() {
	fileSystem := afero.NewMemMapFs()
	_, err := identity.GetAgentID(fileSystem)

	suite.Require().EqualError(err, "open /etc/machine-id: file does not exist")
}
