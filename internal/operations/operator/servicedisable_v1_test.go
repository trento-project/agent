package operator_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/systemd/mocks"
	"github.com/trento-project/agent/internal/operations/operator"
	"github.com/trento-project/agent/pkg/utils"
)

type ServiceDisableOperatorTestSuite struct {
	suite.Suite
	logger            *slog.Logger
	mockSystemd       *mocks.MockSystemd
	mockSystemdLoader *mocks.MockLoader
}

func buildServiceDisableOperator(suite *ServiceDisableOperatorTestSuite) operator.Operator {
	return operator.NewServiceDisable(
		"servicedisableoperator",
		operator.Arguments{},
		"test-op",
		operator.Options[operator.ServiceDisable]{
			BaseOperatorOptions: []operator.BaseOperatorOption{
				operator.WithCustomLogger(suite.logger),
			},
			OperatorOptions: []operator.Option[operator.ServiceDisable]{
				operator.Option[operator.ServiceDisable](operator.WithCustomServiceDisableSystemdLoader(suite.mockSystemdLoader)),
				operator.Option[operator.ServiceDisable](operator.WithServiceToDisable("pacemaker.service")),
			},
		},
	)
}

func TestServiceDisableOperator(t *testing.T) {
	suite.Run(t, new(ServiceDisableOperatorTestSuite))
}

func (suite *ServiceDisableOperatorTestSuite) SetupTest() {
	suite.logger = utils.NewDefaultLogger("info")
	suite.mockSystemd = mocks.NewMockSystemd(suite.T())
	suite.mockSystemdLoader = mocks.NewMockLoader(suite.T())
}

func (suite *ServiceDisableOperatorTestSuite) TestServiceDisableOperatorPlanErrorDbusConnection() {
	ctx := context.Background()

	suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(nil, errors.New("dbus connection error")).
		Once()

	report := buildServiceDisableOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.EqualValues("unable to initialize systemd connector: dbus connection error", report.Error.Message)
}

func (suite *ServiceDisableOperatorTestSuite) TestServiceDisableOperatorPlanErrorIsEnabled() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(suite.mockSystemd, nil).
		Once()

	suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, errors.New("systemd error")).
		Once().
		NotBefore(systemdLoaderCall)

	report := buildServiceDisableOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.EqualValues("failed to check if pacemaker.service service is enabled: systemd error", report.Error.Message)
}

func (suite *ServiceDisableOperatorTestSuite) TestServiceDisableOperatorPlanAlreadyDisabled() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(systemdLoaderCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(isEnabledCall)

	report := buildServiceDisableOperator(suite).Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"enabled":false}`,
		"after":  `{"enabled":false}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *ServiceDisableOperatorTestSuite) TestServiceDisableOperatorCommitErrorDisableFailedRollback() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(true, nil).
		Once().
		NotBefore(systemdLoaderCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(errors.New("systemd disable error")).
		Once().
		NotBefore(isEnabledCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(errors.New("systemd enable error")).
		Once().
		NotBefore(disableCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(enableCall)

	report := buildServiceDisableOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues("systemd enable error\nfailed to disable service pacemaker.service: systemd disable error", report.Error.Message)
}

func (suite *ServiceDisableOperatorTestSuite) TestServiceDisableOperatorCommitErrorDisableSuccessfulRollback() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(true, nil).
		Once().
		NotBefore(systemdLoaderCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(errors.New("systemd disable error")).
		Once().
		NotBefore(isEnabledCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(disableCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(enableCall)

	report := buildServiceDisableOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.EqualValues("failed to disable service pacemaker.service: systemd disable error", report.Error.Message)
}

func (suite *ServiceDisableOperatorTestSuite) TestServiceDisableOperatorVerifyErrorIsDisabledFailedRollback() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(true, nil).
		Once().
		NotBefore(systemdLoaderCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(isEnabledCall)

	verifyIsEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, errors.New("error verifying is disabled")).
		Once().
		NotBefore(disableCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(errors.New("systemd enable error")).
		Once().
		NotBefore(verifyIsEnabledCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(enableCall)

	report := buildServiceDisableOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues("systemd enable error\nfailed to check if service pacemaker.service is enabled: error verifying is disabled", report.Error.Message)
}

func (suite *ServiceDisableOperatorTestSuite) TestServiceDisableOperatorVerifyErrorIsDisabledSuccessfulRollback() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(true, nil).
		Once().
		NotBefore(systemdLoaderCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(isEnabledCall)

	verifyIsEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, errors.New("error verifying is disabled")).
		Once().
		NotBefore(disableCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(verifyIsEnabledCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(enableCall)

	report := buildServiceDisableOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.EqualValues("failed to check if service pacemaker.service is enabled: error verifying is disabled", report.Error.Message)
}

func (suite *ServiceDisableOperatorTestSuite) TestServiceDisableOperatorVerifyNotDisabledFailedRollback() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(true, nil).
		Once().
		NotBefore(systemdLoaderCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(isEnabledCall)

	verifyIsEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(true, nil).
		Once().
		NotBefore(disableCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(errors.New("systemd enable error")).
		Once().
		NotBefore(verifyIsEnabledCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(enableCall)

	report := buildServiceDisableOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues("systemd enable error\nservice pacemaker.service is not disabled", report.Error.Message)
}

func (suite *ServiceDisableOperatorTestSuite) TestServiceDisableOperatorVerifyNotDisabledSuccessfulRollback() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(true, nil).
		Once().
		NotBefore(systemdLoaderCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(isEnabledCall)

	verifyIsEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(true, nil).
		Once().
		NotBefore(disableCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(verifyIsEnabledCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(enableCall)

	report := buildServiceDisableOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.EqualValues("service pacemaker.service is not disabled", report.Error.Message)
}

func (suite *ServiceDisableOperatorTestSuite) TestServiceDisableOperatorSuccess() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, mock.Anything).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(true, nil).
		Once().
		NotBefore(systemdLoaderCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(isEnabledCall)

	verifyIsEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(disableCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(verifyIsEnabledCall)

	report := buildServiceDisableOperator(suite).Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"enabled":true}`,
		"after":  `{"enabled":false}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}
