package cluster_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/cluster"
	"github.com/trento-project/agent/pkg/utils/mocks"
)

const dcNode = "dcNode"

type CmdClientTestSuite struct {
	suite.Suite
}

func TestCmdClient(t *testing.T) {
	suite.Run(t, new(CmdClientTestSuite))
}

func (suite *CmdClientTestSuite) TestGetStateDCError() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCommandExecutor(suite.T())
	mockExecutor.
		On("CombinedOutputContext", mock.Anything, "crmadmin", "-qD").
		Return([]byte(""), errors.New("cluster is not running"))

	cmdClient := cluster.NewCmdClient(mockExecutor, slog.Default())

	_, err := cmdClient.GetState(ctx)
	suite.Error(err, "GetState should return an error")
	suite.EqualError(err, "error getting DC node with crmadmin: cluster is not running")
}

func (suite *CmdClientTestSuite) TestGetStateError() {
	ctx := context.Background()
	dcNodeOutput := []byte(dcNode + "\n")

	mockExecutor := mocks.NewMockCommandExecutor(suite.T())
	mockExecutor.
		On("CombinedOutputContext", mock.Anything, "crmadmin", "-qD").
		Return(dcNodeOutput, nil).
		On("CombinedOutputContext", mock.Anything, "crmadmin", "-qS", dcNode).
		Return([]byte(""), errors.New("error gettings state"))

	cmdClient := cluster.NewCmdClient(mockExecutor, slog.Default())

	_, err := cmdClient.GetState(ctx)
	suite.Error(err, "GetState should return an error")
	suite.EqualError(err, "error getting cluster state with crmadmin: error gettings state")
}

func (suite *CmdClientTestSuite) TestGetStateTimeout() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCommandExecutor(suite.T())
	mockExecutor.
		On("CombinedOutputContext", mock.MatchedBy(func(ctx context.Context) bool {
			_, ok := ctx.Deadline()
			return ok
		}), "crmadmin", "-qD").
		Return(nil, context.DeadlineExceeded)

	cmdClient := cluster.NewCmdClient(mockExecutor, slog.Default())

	_, err := cmdClient.GetState(ctx)
	suite.Error(err, "GetState should return an error")
	suite.EqualError(err, "error getting DC node with crmadmin: context deadline exceeded")
}

func (suite *CmdClientTestSuite) TestGetState() {
	ctx := context.Background()
	dcNodeOutput := []byte(dcNode + "\n")

	mockExecutor := mocks.NewMockCommandExecutor(suite.T())
	mockExecutor.
		On("CombinedOutputContext", mock.Anything, "crmadmin", "-qD").
		Return(dcNodeOutput, nil).
		On("CombinedOutputContext", mock.Anything, "crmadmin", "-qS", dcNode).
		Return([]byte("S_IDLE\n"), nil)

	cmdClient := cluster.NewCmdClient(mockExecutor, slog.Default())

	state, err := cmdClient.GetState(ctx)
	suite.NoError(err, "GetState should not return an error")
	suite.Equal("S_IDLE", state)
}

func (suite *CmdClientTestSuite) TestIsHostOnlineTrue() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCommandExecutor(suite.T())
	mockExecutor.
		On("CombinedOutputContext", ctx, "crm", "status").
		Return([]byte("Online"), nil)

	cmdClient := cluster.NewCmdClient(mockExecutor, slog.Default())

	status := cmdClient.IsHostOnline(ctx)
	suite.True(status, "Cluster should be online")
}

func (suite *CmdClientTestSuite) TestIsHostOnlineFalse() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCommandExecutor(suite.T())
	mockExecutor.
		On("CombinedOutputContext", ctx, "crm", "status").
		Return([]byte("Offline"), errors.New("cluster is not running"))

	cmdClient := cluster.NewCmdClient(mockExecutor, slog.Default())

	status := cmdClient.IsHostOnline(ctx)
	suite.False(status, "Cluster should be offline")
}

func (suite *CmdClientTestSuite) TestIsIdle() {
	ctx := context.Background()

	dcNodeOutput := []byte(dcNode + "\n")

	mockExecutor := mocks.NewMockCommandExecutor(suite.T())
	mockExecutor.
		On("CombinedOutputContext", mock.Anything, "crmadmin", "-qD").
		Return(dcNodeOutput, nil).
		On("CombinedOutputContext", mock.Anything, "crmadmin", "-qS", dcNode).
		Return([]byte("S_IDLE\n"), nil)

	cmdClient := cluster.NewCmdClient(mockExecutor, slog.Default())

	isIdle, err := cmdClient.IsIdle(ctx)
	suite.NoError(err, "IsIdle should not return an error")
	suite.True(isIdle, "Cluster should be idle")
}

func (suite *CmdClientTestSuite) TestIsIdleError() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCommandExecutor(suite.T())
	mockExecutor.
		On("CombinedOutputContext", mock.Anything, "crmadmin", "-qD").
		Return([]byte(""), errors.New("command failed"))

	cmdClient := cluster.NewCmdClient(mockExecutor, slog.Default())

	_, err := cmdClient.IsIdle(ctx)

	suite.Error(err, "IsIdle should return an error")
	suite.EqualError(err, "error getting DC node with crmadmin: command failed")
}

func (suite *CmdClientTestSuite) TestIsIdleDifferentState() {
	ctx := context.Background()

	dcNodeOutput := []byte(dcNode + "\n")

	mockExecutor := mocks.NewMockCommandExecutor(suite.T())
	mockExecutor.
		On("CombinedOutputContext", mock.Anything, "crmadmin", "-qD").
		Return(dcNodeOutput, nil).
		On("CombinedOutputContext", mock.Anything, "crmadmin", "-qS", dcNode).
		Return([]byte("S_TRANSITION_ENGINE"), nil)

	cmdClient := cluster.NewCmdClient(mockExecutor, slog.Default())

	isIdle, err := cmdClient.IsIdle(ctx)
	suite.NoError(err, "IsIdle should not return an error")
	suite.False(isIdle, "Cluster should not be idle")
}

func (suite *CmdClientTestSuite) TestResourceRefresh() {
	ctx := context.Background()
	commandOutput := `Waiting for 1 reply from the controller
... got reply (done)`

	mockExecutor := mocks.NewMockCommandExecutor(suite.T())
	mockExecutor.
		On("CombinedOutputContext", ctx, "crm", "resource", "refresh").
		Return([]byte(commandOutput), nil)

	cmdClient := cluster.NewCmdClient(mockExecutor, slog.Default())

	err := cmdClient.ResourceRefresh(ctx, "", "")
	suite.NoError(err)
}

func (suite *CmdClientTestSuite) TestResourceRefreshWithResource() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCommandExecutor(suite.T())
	mockExecutor.
		On("CombinedOutputContext", ctx, "crm", "resource", "refresh", "my-resource").
		Return([]byte("got reply (done)"), nil)

	cmdClient := cluster.NewCmdClient(mockExecutor, slog.Default())

	err := cmdClient.ResourceRefresh(ctx, "my-resource", "")
	suite.NoError(err)
}

func (suite *CmdClientTestSuite) TestResourceRefreshWithResourceAndNode() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCommandExecutor(suite.T())
	mockExecutor.
		On("CombinedOutputContext", ctx, "crm", "resource", "refresh", "my-resource", "my-node").
		Return([]byte("got reply (done)"), nil)

	cmdClient := cluster.NewCmdClient(mockExecutor, slog.Default())

	err := cmdClient.ResourceRefresh(ctx, "my-resource", "my-node")
	suite.NoError(err)
}

func (suite *CmdClientTestSuite) TestResourceRefreshWithNodeOnlyError() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCommandExecutor(suite.T())

	cmdClient := cluster.NewCmdClient(mockExecutor, slog.Default())

	err := cmdClient.ResourceRefresh(ctx, "", "my-node")
	suite.Error(err)
	suite.EqualError(err, "nodeID cannot be provided without a resourceID")
}

func (suite *CmdClientTestSuite) TestResourceRefreshError() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCommandExecutor(suite.T())
	mockExecutor.
		On("CombinedOutputContext", ctx, "crm", "resource", "refresh").
		Return([]byte("error output"), errors.New("some error"))

	cmdClient := cluster.NewCmdClient(mockExecutor, slog.Default())

	err := cmdClient.ResourceRefresh(ctx, "", "")
	suite.Error(err)
	suite.Contains(err.Error(), "failed to refresh resource")
	suite.Contains(err.Error(), "some error")
	suite.Contains(err.Error(), "error output")
}

func (suite *CmdClientTestSuite) TestResourceRefreshUnexpectedOutputError() {
	ctx := context.Background()

	mockExecutor := mocks.NewMockCommandExecutor(suite.T())
	mockExecutor.
		On("CombinedOutputContext", ctx, "crm", "resource", "refresh").
		Return([]byte("unexpected output"), nil)

	cmdClient := cluster.NewCmdClient(mockExecutor, slog.Default())

	err := cmdClient.ResourceRefresh(ctx, "", "")
	suite.Error(err)
	suite.Contains(err.Error(), "failed to refresh resource, unexpected output")
	suite.Contains(err.Error(), "unexpected output")
}
