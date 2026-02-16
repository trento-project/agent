package operator_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	clusterMocks "github.com/trento-project/agent/internal/core/cluster/mocks"
	"github.com/trento-project/agent/internal/operations/operator"
)

type ClusterResourceRefreshOperatorTestSuite struct {
	suite.Suite
	mockClusterClient *clusterMocks.MockCmdClient
}

func TestClusterResourceRefreshOperator(t *testing.T) {
	suite.Run(t, new(ClusterResourceRefreshOperatorTestSuite))
}

func (suite *ClusterResourceRefreshOperatorTestSuite) SetupTest() {
	suite.mockClusterClient = clusterMocks.NewMockCmdClient(suite.T())
}

func (suite *ClusterResourceRefreshOperatorTestSuite) TestClusterResourceRefreshSuccess() {
	ctx := context.Background()

	suite.mockClusterClient.
		On("IsHostOnline", ctx).Return(true).Once().
		On("IsIdle", ctx).Return(true, nil).Once().
		On("ResourceRefresh", ctx, "", "").Return(nil).Once()

	clusterResourceRefreshOperator := operator.NewClusterResourceRefresh(
		operator.Arguments{},
		"test-op",
		operator.Options[operator.ClusterResourceRefresh]{
			OperatorOptions: []operator.Option[operator.ClusterResourceRefresh]{
				operator.Option[operator.ClusterResourceRefresh](operator.WithCustomClusterResourceRefreshClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterResourceRefreshOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"refreshed":false}`,
		"after":  `{"refreshed":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *ClusterResourceRefreshOperatorTestSuite) TestClusterResourceRefreshSuccessWithResourceID() {
	ctx := context.Background()
	resourceID := "some-resource"

	suite.mockClusterClient.
		On("IsHostOnline", ctx).Return(true).Once().
		On("IsIdle", ctx).Return(true, nil).Once().
		On("ResourceRefresh", ctx, resourceID, "").Return(nil).Once()

	clusterResourceRefreshOperator := operator.NewClusterResourceRefresh(
		operator.Arguments{
			"resource_id": resourceID,
		},
		"test-op",
		operator.Options[operator.ClusterResourceRefresh]{
			OperatorOptions: []operator.Option[operator.ClusterResourceRefresh]{
				operator.Option[operator.ClusterResourceRefresh](operator.WithCustomClusterResourceRefreshClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterResourceRefreshOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"refreshed":false,"resource_id":"some-resource"}`,
		"after":  `{"refreshed":true,"resource_id":"some-resource"}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *ClusterResourceRefreshOperatorTestSuite) TestClusterResourceRefreshSuccessWithResourceIDAndNodeID() {
	ctx := context.Background()
	resourceID := "some-resource"
	nodeID := "some-node"

	suite.mockClusterClient.
		On("IsHostOnline", ctx).Return(true).Once().
		On("IsIdle", ctx).Return(true, nil).Once().
		On("ResourceRefresh", ctx, resourceID, nodeID).Return(nil).Once()

	clusterResourceRefreshOperator := operator.NewClusterResourceRefresh(
		operator.Arguments{
			"resource_id": resourceID,
			"node_id":     nodeID,
		},
		"test-op",
		operator.Options[operator.ClusterResourceRefresh]{
			OperatorOptions: []operator.Option[operator.ClusterResourceRefresh]{
				operator.Option[operator.ClusterResourceRefresh](operator.WithCustomClusterResourceRefreshClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterResourceRefreshOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"refreshed":false,"resource_id":"some-resource","node_id":"some-node"}`,
		"after":  `{"refreshed":true,"resource_id":"some-resource","node_id":"some-node"}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *ClusterResourceRefreshOperatorTestSuite) TestClusterResourceRefreshPlanInvalidArgument() {
	ctx := context.Background()

	clusterResourceRefreshOperator := operator.NewClusterResourceRefresh(
		operator.Arguments{
			"node_id": "some-node",
		},
		"test-op",
		operator.Options[operator.ClusterResourceRefresh]{},
	)

	report := clusterResourceRefreshOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.Equal("node_id cannot be provided without a resource_id", report.Error.Message)
}

func (suite *ClusterResourceRefreshOperatorTestSuite) TestClusterResourceRefreshPlanInvalidArgumentResourceID() {
	ctx := context.Background()

	clusterResourceRefreshOperator := operator.NewClusterResourceRefresh(
		operator.Arguments{
			"resource_id": 10,
		},
		"test-op",
		operator.Options[operator.ClusterResourceRefresh]{},
	)

	report := clusterResourceRefreshOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.Contains(report.Error.Message, "could not parse resource_id argument as string")
}

func (suite *ClusterResourceRefreshOperatorTestSuite) TestClusterResourceRefreshPlanInvalidArgumentNodeID() {
	ctx := context.Background()

	clusterResourceRefreshOperator := operator.NewClusterResourceRefresh(
		operator.Arguments{
			"resource_id": "some-resource",
			"node_id":     10,
		},
		"test-op",
		operator.Options[operator.ClusterResourceRefresh]{},
	)

	report := clusterResourceRefreshOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.Contains(report.Error.Message, "could not parse node_id argument as string")
}

func (suite *ClusterResourceRefreshOperatorTestSuite) TestClusterResourceRefreshPlanClusterOffline() {
	ctx := context.Background()

	suite.mockClusterClient.On("IsHostOnline", ctx).Return(false).Once()

	clusterResourceRefreshOperator := operator.NewClusterResourceRefresh(
		operator.Arguments{},
		"test-op",
		operator.Options[operator.ClusterResourceRefresh]{
			OperatorOptions: []operator.Option[operator.ClusterResourceRefresh]{
				operator.Option[operator.ClusterResourceRefresh](operator.WithCustomClusterResourceRefreshClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterResourceRefreshOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.Equal("cluster is not runnint on host", report.Error.Message)
}

func (suite *ClusterResourceRefreshOperatorTestSuite) TestClusterResourceRefreshCommitNotIdle() {
	ctx := context.Background()

	suite.mockClusterClient.
		On("IsHostOnline", ctx).Return(true).Once().
		On("IsIdle", ctx).Return(false, nil).Once()

	clusterResourceRefreshOperator := operator.NewClusterResourceRefresh(
		operator.Arguments{},
		"test-op",
		operator.Options[operator.ClusterResourceRefresh]{
			OperatorOptions: []operator.Option[operator.ClusterResourceRefresh]{
				operator.Option[operator.ClusterResourceRefresh](operator.WithCustomClusterResourceRefreshClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterResourceRefreshOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.Equal("cluster is not in S_IDLE state", report.Error.Message)
}

func (suite *ClusterResourceRefreshOperatorTestSuite) TestClusterResourceRefreshCommitIsIdleError() {
	ctx := context.Background()
	isIdleError := errors.New("is idle error")

	suite.mockClusterClient.
		On("IsHostOnline", ctx).Return(true).Once().
		On("IsIdle", ctx).Return(false, isIdleError).Once()

	clusterResourceRefreshOperator := operator.NewClusterResourceRefresh(
		operator.Arguments{},
		"test-op",
		operator.Options[operator.ClusterResourceRefresh]{
			OperatorOptions: []operator.Option[operator.ClusterResourceRefresh]{
				operator.Option[operator.ClusterResourceRefresh](operator.WithCustomClusterResourceRefreshClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterResourceRefreshOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.Equal("error checking if cluster is idle: is idle error", report.Error.Message)
}

func (suite *ClusterResourceRefreshOperatorTestSuite) TestClusterResourceRefreshCommitError() {
	ctx := context.Background()
	commitError := errors.New("commit error")

	suite.mockClusterClient.
		On("IsHostOnline", ctx).Return(true).Once().
		On("IsIdle", ctx).Return(true, nil).Once().
		On("ResourceRefresh", ctx, "", "").Return(commitError).Once()

	clusterResourceRefreshOperator := operator.NewClusterResourceRefresh(
		operator.Arguments{},
		"test-op",
		operator.Options[operator.ClusterResourceRefresh]{
			OperatorOptions: []operator.Option[operator.ClusterResourceRefresh]{
				operator.Option[operator.ClusterResourceRefresh](operator.WithCustomClusterResourceRefreshClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterResourceRefreshOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.Equal(commitError.Error(), report.Error.Message)
}
