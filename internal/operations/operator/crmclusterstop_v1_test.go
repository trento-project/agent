package operator_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/cluster/mocks"
	"github.com/trento-project/agent/internal/operations/operator"
)

type CrmClusterStopOperatorTestSuite struct {
	suite.Suite
}

func TestCrmClusterStopOperator(t *testing.T) {
	suite.Run(t, new(CrmClusterStopOperatorTestSuite))
}

func (suite *CrmClusterStopOperatorTestSuite) SetupTest() {
	// Setup code for the test suite can be added here
}

func (suite *CrmClusterStopOperatorTestSuite) TestCrmClusterStopClusterAlreadyOffline() {
	ctx := context.Background()

	mockCmdClient := mocks.NewMockCmdClient(suite.T())
	mockCmdClient.On("IsHostOnline", ctx).Return(false).Once()

	crmClusterStopOperator := operator.NewCrmClusterStop(
		operator.Arguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.Options[operator.CrmClusterStop]{
			OperatorOptions: []operator.Option[operator.CrmClusterStop]{
				operator.Option[operator.CrmClusterStop](operator.WithCustomClusterClientStop(mockCmdClient)),
			},
		},
	)

	report := crmClusterStopOperator.Run(ctx)

	suite.NotNil(report.Success)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(map[string]any{
		"before": `{"stopped":true}`,
		"after":  `{"stopped":true}`,
	}, report.Success.Diff)
}

func (suite *CrmClusterStopOperatorTestSuite) TestCrmClusterStopClusterRollbackFailure() {
	ctx := context.Background()

	mockCmdClient := mocks.NewMockCmdClient(suite.T())
	mockCmdClient.On("IsHostOnline", ctx).Return(true).Once()
	mockCmdClient.On("IsIdle", ctx).Return(false, nil)
	mockCmdClient.On("StartCluster", ctx).Return(errors.New("failed to start cluster"))

	crmClusterStopOperator := operator.NewCrmClusterStop(
		operator.Arguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.Options[operator.CrmClusterStop]{
			OperatorOptions: []operator.Option[operator.CrmClusterStop]{
				operator.Option[operator.CrmClusterStop](operator.WithCustomClusterClientStop(mockCmdClient)),
				operator.Option[operator.CrmClusterStop](operator.WithCustomRetryStop(2, 100*time.Millisecond, 1*time.Second, 1)),
			},
		},
	)

	report := crmClusterStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.NotEmpty(report.Error.Message)
}

func (suite *CrmClusterStopOperatorTestSuite) TestCrmClusterStopClusterRollbackSuccess() {
	ctx := context.Background()

	mockCmdClient := mocks.NewMockCmdClient(suite.T())
	mockCmdClient.On("IsHostOnline", ctx).Return(true).Once()
	mockCmdClient.On("IsIdle", ctx).Return(true, nil)
	mockCmdClient.On("StopCluster", ctx).Return(errors.New("failed to stop cluster"))
	mockCmdClient.On("StartCluster", ctx).Return(nil).Once()

	crmClusterStopOperator := operator.NewCrmClusterStop(
		operator.Arguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.Options[operator.CrmClusterStop]{
			OperatorOptions: []operator.Option[operator.CrmClusterStop]{
				operator.Option[operator.CrmClusterStop](operator.WithCustomClusterClientStop(mockCmdClient)),
			},
		},
	)

	report := crmClusterStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.NotEmpty(report.Error.Message)
}

func (suite *CrmClusterStopOperatorTestSuite) TestCrmClusterStopClusterStartVerifyFailure() {
	ctx := context.Background()

	mockCmdClient := mocks.NewMockCmdClient(suite.T())
	mockCmdClient.On("IsHostOnline", ctx).Return(true).Once()
	mockCmdClient.On("IsIdle", ctx).Return(true, nil)
	mockCmdClient.On("StopCluster", ctx).Return(nil)
	mockCmdClient.On("IsHostOnline", ctx).Return(true)
	mockCmdClient.On("StartCluster", ctx).Return(nil).Once()

	crmClusterStopOperator := operator.NewCrmClusterStop(
		operator.Arguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.Options[operator.CrmClusterStop]{
			OperatorOptions: []operator.Option[operator.CrmClusterStop]{
				operator.Option[operator.CrmClusterStop](operator.WithCustomClusterClientStop(mockCmdClient)),
				operator.Option[operator.CrmClusterStop](operator.WithCustomRetryStop(2, 100*time.Millisecond, 1*time.Second, 2)),
			},
		},
	)

	report := crmClusterStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.NotEmpty(report.Error.Message)
}

func (suite *CrmClusterStopOperatorTestSuite) TestCrmClusterStopVerifySuccess() {
	ctx := context.Background()

	mockCmdClient := mocks.NewMockCmdClient(suite.T())
	mockCmdClient.On("IsHostOnline", ctx).Return(true).Once()
	mockCmdClient.On("IsIdle", ctx).Return(true, nil).Once()
	mockCmdClient.On("StopCluster", ctx).Return(nil).Once()
	mockCmdClient.On("IsHostOnline", ctx).Return(false)

	crmClusterStopOperator := operator.NewCrmClusterStop(
		operator.Arguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.Options[operator.CrmClusterStop]{
			OperatorOptions: []operator.Option[operator.CrmClusterStop]{
				operator.Option[operator.CrmClusterStop](operator.WithCustomClusterClientStop(mockCmdClient)),
			},
		},
	)

	report := crmClusterStopOperator.Run(ctx)

	suite.NotNil(report.Success)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(map[string]any{
		"before": `{"stopped":false}`,
		"after":  `{"stopped":true}`,
	}, report.Success.Diff)
}
