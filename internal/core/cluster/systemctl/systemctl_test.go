package systemctl_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/cluster/systemctl"
	"github.com/trento-project/agent/pkg/utils/mocks"
)

type SystemctlTestSuite struct {
	suite.Suite
}

func TestSystemctlTestSuite(t *testing.T) {
	suite.Run(t, new(SystemctlTestSuite))
}

func (suite *SystemctlTestSuite) TestIsActive() {
	mockCommand := new(mocks.MockCommandExecutor)
	mockCommand.On("Exec", "systemctl", "is-active", "test-service").Return([]byte("active"), nil)

	systemctl := systemctl.NewSystemctl(mockCommand)
	active := systemctl.IsActive("test-service")

	suite.True(active)
}

func (suite *SystemctlTestSuite) TestIsNotActive() {
	mockCommand := new(mocks.MockCommandExecutor)
	mockCommand.On("Exec", "systemctl", "is-active", "test-service").Return([]byte("inactive"), errors.New(""))

	systemctl := systemctl.NewSystemctl(mockCommand)
	active := systemctl.IsActive("test-service")

	suite.False(active)
}
