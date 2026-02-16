package operator_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/systemd/mocks"
	"github.com/trento-project/agent/internal/operations/operator"
	"github.com/trento-project/agent/pkg/utils"
)

type ServiceEnableOperatorTestSuite struct {
	suite.Suite
	logger            *slog.Logger
	mockSystemd       *mocks.MockSystemd
	mockSystemdLoader *mocks.MockLoader
}

func buildServiceEnableOperator(suite *ServiceEnableOperatorTestSuite) operator.Operator {
	return operator.NewServiceEnable(
		"serviceenableoperator",
		operator.Arguments{},
		"test-op",
		operator.Options[operator.ServiceEnable]{
			BaseOperatorOptions: []operator.BaseOperatorOption{
				operator.WithCustomLogger(suite.logger),
			},
			OperatorOptions: []operator.Option[operator.ServiceEnable]{
				operator.Option[operator.ServiceEnable](operator.WithCustomServiceEnableSystemdLoader(suite.mockSystemdLoader)),
				operator.Option[operator.ServiceEnable](operator.WithServiceToEnable("pacemaker.service")),
			},
		},
	)
}

func TestServiceEnableOperator(t *testing.T) {
	suite.Run(t, new(ServiceEnableOperatorTestSuite))
}

func (suite *ServiceEnableOperatorTestSuite) SetupTest() {
	suite.logger = utils.NewDefaultLogger("info")
	suite.mockSystemd = mocks.NewMockSystemd(suite.T())
	suite.mockSystemdLoader = mocks.NewMockLoader(suite.T())
}

func (suite *ServiceEnableOperatorTestSuite) TestServiceEnableOperatorPlanErrorDbusConnection() {
	ctx := context.Background()

	suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(nil, errors.New("dbus connection error")).
		Once()

	report := buildServiceEnableOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.EqualValues("unable to initialize systemd connector: dbus connection error", report.Error.Message)
}

func (suite *ServiceEnableOperatorTestSuite) TestServiceEnableOperatorPlanErrorIsEnabled() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(suite.mockSystemd, nil).
		Once()

	suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, errors.New("systemd error")).
		Once().
		NotBefore(systemdLoaderCall)

	report := buildServiceEnableOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.EqualValues("failed to check if pacemaker.service service is enabled: systemd error", report.Error.Message)
}

func (suite *ServiceEnableOperatorTestSuite) TestServiceEnableOperatorPlanAlreadyEnabled() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(true, nil).
		Once().
		NotBefore(systemdLoaderCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(isEnabledCall)

	report := buildServiceEnableOperator(suite).Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"enabled":true}`,
		"after":  `{"enabled":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *ServiceEnableOperatorTestSuite) TestServiceEnableOperatorCommitErrorEnableFailedRollback() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(systemdLoaderCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(errors.New("systemd enable error")).
		Once().
		NotBefore(isEnabledCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(errors.New("systemd disable error")).
		Once().
		NotBefore(enableCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(disableCall)

	report := buildServiceEnableOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues("systemd disable error\nfailed to enable service pacemaker.service: systemd enable error", report.Error.Message)
}

func (suite *ServiceEnableOperatorTestSuite) TestServiceEnableOperatorCommitErrorEnableSuccessfulRollback() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(systemdLoaderCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(errors.New("systemd enable error")).
		Once().
		NotBefore(isEnabledCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(enableCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(disableCall)

	report := buildServiceEnableOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.EqualValues("failed to enable service pacemaker.service: systemd enable error", report.Error.Message)
}

func (suite *ServiceEnableOperatorTestSuite) TestServiceEnableOperatorVerifyErrorIsEnabledFailedRollback() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(systemdLoaderCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(isEnabledCall)

	verifyIsEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, errors.New("error verifying is enabled")).
		Once().
		NotBefore(enableCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(errors.New("systemd disable error")).
		Once().
		NotBefore(verifyIsEnabledCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(disableCall)

	report := buildServiceEnableOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues("systemd disable error\nfailed to check if service pacemaker.service is enabled: error verifying is enabled", report.Error.Message)
}

func (suite *ServiceEnableOperatorTestSuite) TestServiceEnableOperatorVerifyErrorIsEnabledSuccessfulRollback() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(systemdLoaderCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(isEnabledCall)

	verifyIsEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, errors.New("error verifying is enabled")).
		Once().
		NotBefore(enableCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(verifyIsEnabledCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(disableCall)

	report := buildServiceEnableOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.EqualValues("failed to check if service pacemaker.service is enabled: error verifying is enabled", report.Error.Message)
}

func (suite *ServiceEnableOperatorTestSuite) TestServiceEnableOperatorVerifyNotEnabledFailedRollback() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(systemdLoaderCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(isEnabledCall)

	verifyIsEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(enableCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(errors.New("systemd disable error")).
		Once().
		NotBefore(verifyIsEnabledCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(disableCall)

	report := buildServiceEnableOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues("systemd disable error\nservice pacemaker.service is not enabled", report.Error.Message)
}

func (suite *ServiceEnableOperatorTestSuite) TestServiceEnableOperatorVerifyNotEnabledSuccessfulRollback() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(systemdLoaderCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(isEnabledCall)

	verifyIsEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(enableCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(verifyIsEnabledCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(disableCall)

	report := buildServiceEnableOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.EqualValues("service pacemaker.service is not enabled", report.Error.Message)
}

func (suite *ServiceEnableOperatorTestSuite) TestServiceEnableOperatorSuccess() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(systemdLoaderCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(isEnabledCall)

	verifyIsEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(true, nil).
		Once().
		NotBefore(enableCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(verifyIsEnabledCall)

	report := buildServiceEnableOperator(suite).Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"enabled":false}`,
		"after":  `{"enabled":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}
