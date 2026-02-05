package operator_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	clusterMocks "github.com/trento-project/agent/internal/core/cluster/mocks"
	"github.com/trento-project/agent/internal/operations/operator"
	"github.com/trento-project/agent/pkg/utils/mocks"
)

const fakeID = "some-id"

type ClusterMaintenanceChangeOperatorTestSuite struct {
	suite.Suite
	mockCmdExecutor   *mocks.MockCommandExecutor
	mockClusterClient *clusterMocks.MockCmdClient
}

func TestClusterMaintenanceChangeOperator(t *testing.T) {
	suite.Run(t, new(ClusterMaintenanceChangeOperatorTestSuite))
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) SetupTest() {
	suite.mockCmdExecutor = mocks.NewMockCommandExecutor(suite.T())
	suite.mockClusterClient = clusterMocks.NewMockCmdClient(suite.T())
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangeSuccessOn() {
	ctx := context.Background()

	suite.mockClusterClient.On("IsHostOnline", ctx).Return(true)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("false"), nil).Once()

	suite.mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"maintenance",
		"on",
	).Return([]byte("ok"), nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("true"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "{\"maintenance\":false}",
		"after":  "{\"maintenance\":true}",
	}

	suite.Nil(report.Error)
	suite.Equal(report.Success.LastPhase, operator.VERIFY)
	suite.EqualValues(report.Success.Diff, expectedDiff)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangeSuccessOff() {
	ctx := context.Background()

	suite.mockClusterClient.On("IsHostOnline", ctx).Return(true)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("true"), nil).Once()

	suite.mockClusterClient.
		On("IsIdle", ctx).Return(true, nil).
		On("ResourceRefresh", ctx, "", "").Return(nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"maintenance",
		"off",
	).Return([]byte("ok"), nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("false"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": false,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "{\"maintenance\":true}",
		"after":  "{\"maintenance\":false}",
	}

	suite.Nil(report.Error)
	suite.Equal(report.Success.LastPhase, operator.VERIFY)
	suite.EqualValues(report.Success.Diff, expectedDiff)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangeResourceSuccess() {
	ctx := context.Background()
	resourceID := fakeID

	suite.mockClusterClient.On("IsHostOnline", ctx).Return(true)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"resource",
		"meta",
		resourceID,
		"show",
		"maintenance",
	).Return([]byte("false"), nil).Once()

	suite.mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"maintenance",
		"on",
		resourceID,
	).Return([]byte("ok"), nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"resource",
		"meta",
		resourceID,
		"show",
		"maintenance",
	).Return([]byte("true"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
			"resource_id": resourceID,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "{\"maintenance\":false,\"resource_id\":\"some-id\"}",
		"after":  "{\"maintenance\":true,\"resource_id\":\"some-id\"}",
	}

	suite.Nil(report.Error)
	suite.Equal(report.Success.LastPhase, operator.VERIFY)
	suite.EqualValues(report.Success.Diff, expectedDiff)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangeResourceWithIsManagedSuccess() {
	ctx := context.Background()
	resourceID := fakeID

	suite.mockClusterClient.On("IsHostOnline", ctx).Return(true)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"resource",
		"meta",
		resourceID,
		"show",
		"maintenance",
	).Return([]byte("not found"), nil)

	// is-managed has the reverse boolean logic than `maintenance`
	// so is-managed=true means maintenance=false
	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"resource",
		"meta",
		resourceID,
		"show",
		"is-managed",
	).Return([]byte("true"), nil).Once()

	suite.mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"maintenance",
		"on",
		resourceID,
	).Return([]byte("ok"), nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"resource",
		"meta",
		resourceID,
		"show",
		"is-managed",
	).Return([]byte("false"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
			"resource_id": resourceID,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "{\"maintenance\":false,\"resource_id\":\"some-id\"}",
		"after":  "{\"maintenance\":true,\"resource_id\":\"some-id\"}",
	}

	suite.Nil(report.Error)
	suite.Equal(report.Success.LastPhase, operator.VERIFY)
	suite.EqualValues(report.Success.Diff, expectedDiff)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangeResourceDefaultSuccess() {
	ctx := context.Background()
	resourceID := fakeID

	suite.mockClusterClient.On("IsHostOnline", ctx).Return(true)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"resource",
		"meta",
		resourceID,
		"show",
		"maintenance",
	).Return([]byte("not found"), nil).Once()

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"resource",
		"meta",
		resourceID,
		"show",
		"is-managed",
	).Return([]byte("not found"), nil)

	suite.mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"maintenance",
		"on",
		resourceID,
	).Return([]byte("ok"), nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"resource",
		"meta",
		resourceID,
		"show",
		"maintenance",
	).Return([]byte("true"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
			"resource_id": resourceID,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "{\"maintenance\":false,\"resource_id\":\"some-id\"}",
		"after":  "{\"maintenance\":true,\"resource_id\":\"some-id\"}",
	}

	suite.Nil(report.Error)
	suite.Equal(report.Success.LastPhase, operator.VERIFY)
	suite.EqualValues(report.Success.Diff, expectedDiff)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangeNodeSuccessOn() {
	ctx := context.Background()
	nodeID := fakeID

	suite.mockClusterClient.On("IsHostOnline", ctx).Return(true)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"node",
		"attribute",
		nodeID,
		"show",
		"maintenance",
	).Return([]byte("scope=nodes  name=maintenance value=off"), nil).Once()

	suite.mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"--force",
		"node",
		"maintenance",
		nodeID,
	).Return([]byte("ok"), nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"node",
		"attribute",
		nodeID,
		"show",
		"maintenance",
	).Return([]byte("scope=nodes  name=maintenance value=true"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
			"node_id":     nodeID,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "{\"maintenance\":false,\"node_id\":\"some-id\"}",
		"after":  "{\"maintenance\":true,\"node_id\":\"some-id\"}",
	}

	suite.Nil(report.Error)
	suite.Equal(report.Success.LastPhase, operator.VERIFY)
	suite.EqualValues(report.Success.Diff, expectedDiff)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangeNodeSuccessOnWithoutPreviousState() {
	ctx := context.Background()
	nodeID := fakeID

	suite.mockClusterClient.On("IsHostOnline", ctx).Return(true)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"node",
		"attribute",
		nodeID,
		"show",
		"maintenance",
	).Return(
		[]byte("scope=nodes  name=maintenance value=(null)"),
		errors.New("error getting node state"),
	).Once()

	suite.mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"--force",
		"node",
		"maintenance",
		nodeID,
	).Return([]byte("ok"), nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"node",
		"attribute",
		nodeID,
		"show",
		"maintenance",
	).Return([]byte("scope=nodes  name=maintenance value=true"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
			"node_id":     nodeID,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "{\"maintenance\":false,\"node_id\":\"some-id\"}",
		"after":  "{\"maintenance\":true,\"node_id\":\"some-id\"}",
	}

	suite.Nil(report.Error)
	suite.Equal(report.Success.LastPhase, operator.VERIFY)
	suite.EqualValues(report.Success.Diff, expectedDiff)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangeNodeSuccessOff() {
	ctx := context.Background()
	nodeID := fakeID

	suite.mockClusterClient.On("IsHostOnline", ctx).Return(true)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"node",
		"attribute",
		nodeID,
		"show",
		"maintenance",
	).Return([]byte("scope=nodes  name=maintenance value=true"), nil).Once()

	suite.mockClusterClient.
		On("IsIdle", ctx).Return(true, nil).
		On("ResourceRefresh", ctx, "", "").Return(nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"--force",
		"node",
		"ready",
		nodeID,
	).Return([]byte("ok"), nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"node",
		"attribute",
		nodeID,
		"show",
		"maintenance",
	).Return([]byte("scope=nodes  name=maintenance value=off"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": false,
			"node_id":     nodeID,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "{\"maintenance\":true,\"node_id\":\"some-id\"}",
		"after":  "{\"maintenance\":false,\"node_id\":\"some-id\"}",
	}

	suite.Nil(report.Error)
	suite.Equal(report.Success.LastPhase, operator.VERIFY)
	suite.EqualValues(report.Success.Diff, expectedDiff)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangeMissingArgument() {
	ctx := context.Background()

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(report.Error.ErrorPhase, operator.PLAN)
	suite.EqualValues("argument maintenance not provided, could not use the operator", report.Error.Message)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangeInvalidArgument() {
	ctx := context.Background()

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": "on",
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(report.Error.ErrorPhase, operator.PLAN)
	suite.EqualValues("could not parse maintenance argument as bool, argument provided: on", report.Error.Message)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangeInvalidResourceIDArgument() {
	ctx := context.Background()

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
			"resource_id": 1,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(report.Error.ErrorPhase, operator.PLAN)
	suite.EqualValues("could not parse resource_id argument as string, argument provided: 1", report.Error.Message)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangeInvalidNodeIDArgument() {
	ctx := context.Background()

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
			"node_id":     1,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(report.Error.ErrorPhase, operator.PLAN)
	suite.EqualValues("could not parse node_id argument as string, argument provided: 1", report.Error.Message)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangeMutuallyExclusiveArgument() {
	ctx := context.Background()

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
			"resource_id": "some-resource",
			"node_id":     "some-node",
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(report.Error.ErrorPhase, operator.PLAN)
	suite.EqualValues("resource_id and node_id arguments are mutually exclusive, use only one of them", report.Error.Message)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangePlanClusterNotFound() {
	ctx := context.Background()

	suite.mockClusterClient.On("IsHostOnline", ctx).Return(false)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(report.Error.ErrorPhase, operator.PLAN)
	suite.EqualValues("cluster is not runnint on host", report.Error.Message)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangePlanGetMaintenanceError() {
	ctx := context.Background()

	suite.mockClusterClient.On("IsHostOnline", ctx).Return(true)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("error"), errors.New("cannot get state"))

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(report.Error.ErrorPhase, operator.PLAN)
	suite.EqualValues("error getting maintenance-mode: cannot get state", report.Error.Message)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangePlanEmptyMaintenanceState() {
	ctx := context.Background()

	suite.mockClusterClient.On("IsHostOnline", ctx).Return(true)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte(""), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(report.Error.ErrorPhase, operator.PLAN)
	suite.EqualValues("error decoding maintenance-mode attribute: empty command output", report.Error.Message)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangePlanNodeNotFound() {
	ctx := context.Background()
	nodeID := fakeID

	suite.mockClusterClient.On("IsHostOnline", ctx).Return(true)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"node",
		"attribute",
		nodeID,
		"show",
		"maintenance",
	).Return([]byte("Could not map name=some-id to a UUID"), errors.New("error getting node"))

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
			"node_id":     nodeID,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(report.Error.ErrorPhase, operator.PLAN)
	suite.EqualValues("error getting node maintenance attribute: error getting node", report.Error.Message)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangeCommitAlreadyApplied() {
	ctx := context.Background()

	suite.mockClusterClient.On("IsHostOnline", ctx).Return(true)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("true"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": "{\"maintenance\":true}",
		"after":  "{\"maintenance\":true}",
	}

	suite.Nil(report.Error)
	suite.Equal(report.Success.LastPhase, operator.PLAN)
	suite.EqualValues(report.Success.Diff, expectedDiff)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangeCommitNotIdle() {
	ctx := context.Background()

	suite.mockClusterClient.On("IsHostOnline", ctx).Return(true)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("false"), nil)

	suite.mockClusterClient.On("IsIdle", ctx).Return(false, nil).Once()
	suite.mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"maintenance",
		"off",
	).Return([]byte("ok"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(report.Error.ErrorPhase, operator.COMMIT)
	suite.EqualValues("cluster is not in S_IDLE state", report.Error.Message)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangeVerifyError() {
	ctx := context.Background()

	suite.mockClusterClient.On("IsHostOnline", ctx).Return(true)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("false"), nil).Once()

	suite.mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"maintenance",
		"on",
	).Return([]byte("ok"), nil).Once()

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("false"), nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"maintenance",
		"off",
	).Return([]byte("ok"), nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(report.Error.ErrorPhase, operator.VERIFY)
	suite.EqualValues("verify cluster maintenance failed, the maintenance value true was not set in commit phase", report.Error.Message)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangeRollbackNotIdle() {
	ctx := context.Background()

	suite.mockClusterClient.On("IsHostOnline", ctx).Return(true)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("false"), nil)

	suite.mockClusterClient.On("IsIdle", ctx).Return(true, nil).Once()

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"maintenance",
		"on",
	).Return([]byte("error"), errors.New("error changing"))

	suite.mockClusterClient.On("IsIdle", ctx).Return(false, nil)

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(report.Error.ErrorPhase, operator.ROLLBACK)
	suite.EqualValues("cluster is not in S_IDLE state\nerror updating maintenance state: error changing", report.Error.Message)
}

func (suite *ClusterMaintenanceChangeOperatorTestSuite) TestClusterMaintenanceChangeRollbackErrorReverting() {
	ctx := context.Background()

	suite.mockClusterClient.On("IsHostOnline", ctx).Return(true)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"configure",
		"get_property",
		"-t",
		"maintenance-mode",
	).Return([]byte("false"), nil)

	suite.mockClusterClient.On("IsIdle", ctx).Return(true, nil)

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"maintenance",
		"on",
	).Return([]byte("error"), errors.New("error changing"))

	suite.mockCmdExecutor.On(
		"CombinedOutputContext",
		ctx,
		"crm",
		"maintenance",
		"off",
	).Return([]byte("error"), errors.New("error reverting"))

	clusterMaintenanceChangeOperator := operator.NewClusterMaintenanceChange(
		operator.Arguments{
			"maintenance": true,
		},
		"test-op",
		operator.Options[operator.ClusterMaintenanceChange]{
			OperatorOptions: []operator.Option[operator.ClusterMaintenanceChange]{
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceExecutor(suite.mockCmdExecutor)),
				operator.Option[operator.ClusterMaintenanceChange](operator.WithCustomClusterMaintenanceClient(suite.mockClusterClient)),
			},
		},
	)

	report := clusterMaintenanceChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(report.Error.ErrorPhase, operator.ROLLBACK)
	suite.EqualValues("error rolling back maintenance state: error reverting\nerror updating maintenance state: error changing", report.Error.Message)
}
