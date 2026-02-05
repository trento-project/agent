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

type CrmClusterStartOperatorTestSuite struct {
	suite.Suite
}

func TestCrmClusterStartOperator(t *testing.T) {
	suite.Run(t, new(CrmClusterStartOperatorTestSuite))
}

func (suite *CrmClusterStartOperatorTestSuite) SetupTest() {
	// Setup code for the test suite can be added here
}

func (suite *CrmClusterStartOperatorTestSuite) TestCrmClusterStartClusterAlreadyOnline() {
	ctx := context.Background()

	mockCmdClient := mocks.NewMockCmdClient(suite.T())
	mockCmdClient.On("IsHostOnline", ctx).Return(true).Once()

	crmClusterStartOperator := operator.NewCrmClusterStart(
		operator.Arguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.Options[operator.CrmClusterStart]{
			OperatorOptions: []operator.Option[operator.CrmClusterStart]{
				operator.Option[operator.CrmClusterStart](operator.WithCustomClusterClient(mockCmdClient)),
			},
		},
	)

	report := crmClusterStartOperator.Run(ctx)

	suite.NotNil(report.Success)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(map[string]any{
		"before": `{"started":true}`,
		"after":  `{"started":true}`,
	}, report.Success.Diff)
}

func (suite *CrmClusterStartOperatorTestSuite) TestCrmClusterStartClusterRollbackFailure() {
	ctx := context.Background()

	mockCmdClient := mocks.NewMockCmdClient(suite.T())
	mockCmdClient.On("IsHostOnline", ctx).Return(false).Once()
	mockCmdClient.On("StartCluster", ctx).Return(errors.New("failed to start cluster")).Once()
	mockCmdClient.On("IsIdle", ctx).Return(true, nil)
	mockCmdClient.On("StopCluster", ctx).Return(errors.New("failed to stop cluster"))

	crmClusterStartOperator := operator.NewCrmClusterStart(
		operator.Arguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.Options[operator.CrmClusterStart]{
			OperatorOptions: []operator.Option[operator.CrmClusterStart]{
				operator.Option[operator.CrmClusterStart](operator.WithCustomClusterClient(mockCmdClient)),
				operator.Option[operator.CrmClusterStart](operator.WithCustomRetry(2, 100*time.Millisecond, 1*time.Second, 1)),
			},
		},
	)

	report := crmClusterStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.NotEmpty(report.Error.Message)
}

func (suite *CrmClusterStartOperatorTestSuite) TestCrmClusterStartClusterRollbackFailureNotIdle() {
	ctx := context.Background()

	mockCmdClient := mocks.NewMockCmdClient(suite.T())
	mockCmdClient.On("IsHostOnline", ctx).Return(false).Once()
	mockCmdClient.On("StartCluster", ctx).Return(errors.New("failed to start cluster")).Once()
	mockCmdClient.On("IsIdle", ctx).Return(false, nil)

	crmClusterStartOperator := operator.NewCrmClusterStart(
		operator.Arguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.Options[operator.CrmClusterStart]{
			OperatorOptions: []operator.Option[operator.CrmClusterStart]{
				operator.Option[operator.CrmClusterStart](operator.WithCustomClusterClient(mockCmdClient)),
				operator.Option[operator.CrmClusterStart](operator.WithCustomRetry(2, 100*time.Millisecond, 1*time.Second, 2)),
			},
		},
	)

	report := crmClusterStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.NotEmpty(report.Error.Message)
}

func (suite *CrmClusterStartOperatorTestSuite) TestCrmClusterStartClusterRollbackSuccess() {
	ctx := context.Background()

	mockCmdClient := mocks.NewMockCmdClient(suite.T())
	mockCmdClient.On("IsIdle", ctx).Return(true, nil)
	mockCmdClient.On("IsHostOnline", ctx).Return(false).Once()
	mockCmdClient.On("StartCluster", ctx).Return(errors.New("failed to start cluster")).Once()
	mockCmdClient.On("StopCluster", ctx).Return(nil).Once()

	crmClusterStartOperator := operator.NewCrmClusterStart(
		operator.Arguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.Options[operator.CrmClusterStart]{
			OperatorOptions: []operator.Option[operator.CrmClusterStart]{
				operator.Option[operator.CrmClusterStart](operator.WithCustomClusterClient(mockCmdClient)),
			},
		},
	)

	report := crmClusterStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.NotEmpty(report.Error.Message)
}

func (suite *CrmClusterStartOperatorTestSuite) TestCrmClusterStartClusterStartVerifyFailure() {
	ctx := context.Background()

	mockCmdClient := mocks.NewMockCmdClient(suite.T())
	mockCmdClient.On("IsIdle", ctx).Return(true, nil)
	mockCmdClient.On("IsHostOnline", ctx).Return(false).Once()
	mockCmdClient.On("StartCluster", ctx).Return(nil).Once()
	mockCmdClient.On("IsHostOnline", ctx).Return(false)
	mockCmdClient.On("StopCluster", ctx).Return(nil).Once()

	crmClusterStartOperator := operator.NewCrmClusterStart(
		operator.Arguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.Options[operator.CrmClusterStart]{
			OperatorOptions: []operator.Option[operator.CrmClusterStart]{
				operator.Option[operator.CrmClusterStart](operator.WithCustomClusterClient(mockCmdClient)),
				operator.Option[operator.CrmClusterStart](operator.WithCustomRetry(2, 100*time.Millisecond, 1*time.Second, 2)),
			},
		},
	)

	report := crmClusterStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.NotEmpty(report.Error.Message)
}

func (suite *CrmClusterStartOperatorTestSuite) TestCrmClusterStartClusterStartVerifySuccess() {
	ctx := context.Background()

	mockCmdClient := mocks.NewMockCmdClient(suite.T())
	mockCmdClient.On("IsHostOnline", ctx).Return(false).Once()
	mockCmdClient.On("StartCluster", ctx).Return(nil).Once()
	mockCmdClient.On("IsHostOnline", ctx).Return(true)

	crmClusterStartOperator := operator.NewCrmClusterStart(
		operator.Arguments{
			"cluster_id": "test-cluster-id",
		},
		"test-op",
		operator.Options[operator.CrmClusterStart]{
			OperatorOptions: []operator.Option[operator.CrmClusterStart]{
				operator.Option[operator.CrmClusterStart](operator.WithCustomClusterClient(mockCmdClient)),
			},
		},
	)

	report := crmClusterStartOperator.Run(ctx)

	suite.NotNil(report.Success)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(map[string]any{
		"before": `{"started":false}`,
		"after":  `{"started":true}`,
	}, report.Success.Diff)
}
