package operator_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/dbus"
	dbusMocks "github.com/trento-project/agent/internal/core/dbus/mocks"
	"github.com/trento-project/agent/internal/operations/operator"
	"github.com/trento-project/agent/pkg/utils"
	utilsMocks "github.com/trento-project/agent/pkg/utils/mocks"

	baseDbus "github.com/coreos/go-systemd/v22/dbus"
)

type HostRebootOperatorTestSuite struct {
	suite.Suite
	logger *slog.Logger
}

func buildHostRebootOperator(suite *HostRebootOperatorTestSuite,
	mockCmdExecutor *utilsMocks.MockCommandExecutor,
	mockDbusConnector *dbusMocks.MockConnector,
) *operator.Executor {
	return operator.NewHostReboot(
		operator.Arguments{},
		"test-op",
		operator.Options[operator.HostReboot]{
			BaseOperatorOptions: []operator.BaseOperatorOption{
				operator.WithCustomLogger(suite.logger),
			},
			OperatorOptions: []operator.Option[operator.HostReboot]{
				operator.Option[operator.HostReboot](operator.WithCustomHostRebootExecutor(mockCmdExecutor)),
				operator.Option[operator.HostReboot](operator.WithStaticDbusConnector(mockDbusConnector)),
			},
		},
	)
}

func TestHostRebootOperator(t *testing.T) {
	suite.Run(t, new(HostRebootOperatorTestSuite))
}

func (suite *HostRebootOperatorTestSuite) SetupTest() {
	suite.logger = utils.NewDefaultLogger("info")
}

func (suite *HostRebootOperatorTestSuite) TestHostRebootOperatorSuccess() {
	ctx := context.Background()
	mockCmdExecutor := utilsMocks.NewMockCommandExecutor(suite.T())
	mockDbusConnector := dbusMocks.NewMockConnector(suite.T())

	// Plan phase - check if reboot is already scheduled (it's not)
	mockDbusConnector.On("ListJobsContext", ctx).
		Return([]baseDbus.JobStatus{}, nil).
		Once()

	mockDbusConnector.On("ListUnitsContext", ctx).
		Return([]baseDbus.UnitStatus{}, nil).
		Once()

	mockDbusConnector.On("Close").
		Return().
		Once()

	// Check for active shutdown processes (none found)
	mockCmdExecutor.On("CombinedOutputContext", ctx, "pgrep", "-f", "shutdown").
		Return([]byte(""), errors.New("no process found")).
		Once()

	mockCmdExecutor.On("CombinedOutputContext", ctx, "pgrep", "-f", "systemd-shutdown").
		Return([]byte(""), errors.New("no process found")).
		Once()

	mockCmdExecutor.On("CombinedOutputContext", ctx, "test", "-f", "/run/systemd/shutdown/scheduled").
		Return([]byte(""), errors.New("file not found")).
		Once()

	// Commit phase - schedule reboot
	mockCmdExecutor.On("CombinedOutputContext", ctx, "shutdown", "-r", "+1", "Host reboot scheduled by automation").
		Return([]byte("Reboot scheduled"), nil).
		Once()

	// Verify phase - check that reboot is now scheduled
	rebootJob := baseDbus.JobStatus{
		Id:      1,
		JobType: "start",
		Status:  "waiting",
		Unit:    "reboot.target",
		JobPath: "/org/freedesktop/systemd1/job/1",
	}

	mockDbusConnector.On("ListJobsContext", ctx).
		Return([]baseDbus.JobStatus{rebootJob}, nil).
		Once()

	mockDbusConnector.On("Close").
		Return().
		Once()

	report := buildHostRebootOperator(suite, mockCmdExecutor, mockDbusConnector).Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"scheduled":false}`,
		"after":  `{"scheduled":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *HostRebootOperatorTestSuite) TestHostRebootOperatorAlreadyScheduled() {
	ctx := context.Background()

	mockCmdExecutor := utilsMocks.NewMockCommandExecutor(suite.T())
	mockDbusConnector := dbusMocks.NewMockConnector(suite.T())

	// Plan phase - reboot is already scheduled
	shutdownJob := baseDbus.JobStatus{
		Id:      1,
		JobType: "start",
		Status:  "waiting",
		Unit:    "shutdown.target",
	}

	mockDbusConnector.On("ListJobsContext", ctx).
		Return([]baseDbus.JobStatus{shutdownJob}, nil).
		Once()

	mockDbusConnector.On("Close").
		Return().
		Once()
	report := buildHostRebootOperator(suite, mockCmdExecutor, mockDbusConnector).Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"scheduled":true}`,
		"after":  `{"scheduled":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *HostRebootOperatorTestSuite) TestHostRebootOperatorDbusConnectionError() {
	ctx := context.Background()

	failingConstructor := func(_ context.Context) (dbus.Connector, error) {
		return nil, errors.New("dbus constructor failure")
	}

	report := operator.NewHostReboot(
		operator.Arguments{},
		"test-op",
		operator.Options[operator.HostReboot]{
			BaseOperatorOptions: []operator.BaseOperatorOption{
				operator.WithCustomLogger(suite.logger),
			},
			OperatorOptions: []operator.Option[operator.HostReboot]{
				operator.Option[operator.HostReboot](operator.WithCustomDbusConstructor(failingConstructor)),
			},
		},
	).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.Contains(report.Error.Message, "error checking if reboot is scheduled")
	suite.Contains(report.Error.Message, "failed to connect to systemd D-Bus")
}

func (suite *HostRebootOperatorTestSuite) TestHostRebootOperatorListJobsError() {
	ctx := context.Background()

	mockCmdExecutor := utilsMocks.NewMockCommandExecutor(suite.T())
	mockDbusConnector := dbusMocks.NewMockConnector(suite.T())

	// Plan phase - ListJobs returns error
	mockDbusConnector.On("ListJobsContext", ctx).
		Return([]baseDbus.JobStatus{}, errors.New("D-Bus error")).
		Once()

	mockDbusConnector.On("Close").
		Return().
		Once()
	report := buildHostRebootOperator(suite, mockCmdExecutor, mockDbusConnector).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.Contains(report.Error.Message, "error checking if reboot is scheduled")
	suite.Contains(report.Error.Message, "failed to list systemd jobs")
}

func (suite *HostRebootOperatorTestSuite) TestHostRebootOperatorListUnitsError() {
	ctx := context.Background()

	mockCmdExecutor := utilsMocks.NewMockCommandExecutor(suite.T())
	mockDbusConnector := dbusMocks.NewMockConnector(suite.T())

	// Plan phase - ListJobs succeeds but ListUnits returns error
	mockDbusConnector.On("ListJobsContext", ctx).
		Return([]baseDbus.JobStatus{}, nil).
		Once()

	mockDbusConnector.On("ListUnitsContext", ctx).
		Return([]baseDbus.UnitStatus{}, errors.New("D-Bus units error")).
		Once()
	mockDbusConnector.On("Close").
		Return().
		Once()
	report := buildHostRebootOperator(suite, mockCmdExecutor, mockDbusConnector).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.Contains(report.Error.Message, "error checking if reboot is scheduled")
	suite.Contains(report.Error.Message, "failed to list systemd units")
}

func (suite *HostRebootOperatorTestSuite) TestHostRebootOperatorCommitError() {
	ctx := context.Background()

	mockCmdExecutor := utilsMocks.NewMockCommandExecutor(suite.T())
	mockDbusConnector := dbusMocks.NewMockConnector(suite.T())

	// Plan phase - no reboot scheduled
	mockDbusConnector.On("ListJobsContext", ctx).
		Return([]baseDbus.JobStatus{}, nil).
		Once()

	mockDbusConnector.On("ListUnitsContext", ctx).
		Return([]baseDbus.UnitStatus{}, nil).
		Once()

	mockDbusConnector.On("Close").
		Return().
		Once()

	// Check for active shutdown processes (none found)
	mockCmdExecutor.On("CombinedOutputContext", ctx, "pgrep", "-f", "shutdown").
		Return([]byte(""), errors.New("no process found")).
		Once()

	mockCmdExecutor.On("CombinedOutputContext", ctx, "pgrep", "-f", "systemd-shutdown").
		Return([]byte(""), errors.New("no process found")).
		Once()

	mockCmdExecutor.On("CombinedOutputContext", ctx, "test", "-f", "/run/systemd/shutdown/scheduled").
		Return([]byte(""), errors.New("file not found")).
		Once()

	// Commit phase - shutdown command fails
	mockCmdExecutor.On("CombinedOutputContext", ctx, "shutdown", "-r", "+1", "Host reboot scheduled by automation").
		Return([]byte(""), errors.New("shutdown command failed")).
		Once()

	// Rollback phase - cancel shutdown
	mockCmdExecutor.On("CombinedOutputContext", ctx, "shutdown", "-c").
		Return([]byte("Shutdown cancelled"), nil).
		Once()

	report := buildHostRebootOperator(suite, mockCmdExecutor, mockDbusConnector).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.Contains(report.Error.Message, "shutdown command failed")
}

func (suite *HostRebootOperatorTestSuite) TestHostRebootOperatorVerifyError() {
	ctx := context.Background()

	mockCmdExecutor := utilsMocks.NewMockCommandExecutor(suite.T())
	mockDbusConnector := dbusMocks.NewMockConnector(suite.T())

	// Plan phase - no reboot scheduled
	mockDbusConnector.On("ListJobsContext", ctx).
		Return([]baseDbus.JobStatus{}, nil).
		Once()

	mockDbusConnector.On("ListUnitsContext", ctx).
		Return([]baseDbus.UnitStatus{}, nil).
		Once()

	// Check for active shutdown processes (none found)
	mockCmdExecutor.On("CombinedOutputContext", ctx, "pgrep", "-f", "shutdown").
		Return([]byte(""), errors.New("no process found")).
		Once()

	mockCmdExecutor.On("CombinedOutputContext", ctx, "pgrep", "-f", "systemd-shutdown").
		Return([]byte(""), errors.New("no process found")).
		Once()
	mockCmdExecutor.On("CombinedOutputContext", ctx, "test", "-f", "/run/systemd/shutdown/scheduled").
		Return([]byte(""), errors.New("file not found")).
		Once()

	// Commit phase - schedule reboot
	mockCmdExecutor.On("CombinedOutputContext", ctx, "shutdown", "-r", "+1", "Host reboot scheduled by automation").
		Return([]byte("Reboot scheduled"), nil).
		Once()

	// Verify phase - still no reboot found (verification fails)
	mockDbusConnector.On("ListJobsContext", ctx).
		Return([]baseDbus.JobStatus{}, nil).
		Once()

	mockDbusConnector.On("ListUnitsContext", ctx).
		Return([]baseDbus.UnitStatus{}, nil).
		Once()
	mockDbusConnector.On("Close").
		Return().
		Once()

	// Check for active shutdown processes during verify (none found)
	mockCmdExecutor.On("CombinedOutputContext", ctx, "pgrep", "-f", "shutdown").
		Return([]byte(""), errors.New("no process found")).
		Once()

	mockCmdExecutor.On("CombinedOutputContext", ctx, "pgrep", "-f", "systemd-shutdown").
		Return([]byte(""), errors.New("no process found")).
		Once()

	mockCmdExecutor.On("CombinedOutputContext", ctx, "test", "-f", "/run/systemd/shutdown/scheduled").
		Return([]byte(""), errors.New("file not found")).
		Once()

	// Rollback phase
	mockCmdExecutor.On("CombinedOutputContext", ctx, "shutdown", "-c").
		Return([]byte("Shutdown cancelled"), nil).
		Once()

	mockDbusConnector.On("Close").
		Return().
		Once()

	report := buildHostRebootOperator(suite, mockCmdExecutor, mockDbusConnector).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.Contains(report.Error.Message, "reboot verification failed: no scheduled reboot found")
}

func (suite *HostRebootOperatorTestSuite) TestHostRebootOperatorRollbackError() {
	ctx := context.Background()

	mockCmdExecutor := utilsMocks.NewMockCommandExecutor(suite.T())
	mockDbusConnector := dbusMocks.NewMockConnector(suite.T())

	// Plan phase - no reboot scheduled
	mockDbusConnector.On("ListJobsContext", ctx).
		Return([]baseDbus.JobStatus{}, nil).
		Once()

	mockDbusConnector.On("ListUnitsContext", ctx).
		Return([]baseDbus.UnitStatus{}, nil).Once()

	// Check for active shutdown processes (none found)
	mockCmdExecutor.On("CombinedOutputContext", ctx, "pgrep", "-f", "shutdown").
		Return([]byte(""), errors.New("no process found")).Once()

	mockCmdExecutor.On("CombinedOutputContext", ctx, "pgrep", "-f", "systemd-shutdown").
		Return([]byte(""), errors.New("no process found")).Once()

	mockCmdExecutor.On("CombinedOutputContext", ctx, "test", "-f", "/run/systemd/shutdown/scheduled").
		Return([]byte(""), errors.New("file not found")).Once()

	// Commit phase - shutdown command fails
	mockCmdExecutor.On("CombinedOutputContext", ctx, "shutdown", "-r", "+1", "Host reboot scheduled by automation").
		Return([]byte(""), errors.New("shutdown command failed")).Once()

	// Rollback phase - cancel shutdown also fails
	mockCmdExecutor.On("CombinedOutputContext", ctx, "shutdown", "-c").
		Return([]byte(""), errors.New("cancel command failed")).Once()

	mockDbusConnector.On("Close").Return().Once()

	report := buildHostRebootOperator(suite, mockCmdExecutor, mockDbusConnector).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.Contains(report.Error.Message, "cancel command failed")
}

func (suite *HostRebootOperatorTestSuite) TestHostRebootOperatorShutdownTimerDetection() {
	ctx := context.Background()

	mockCmdExecutor := utilsMocks.NewMockCommandExecutor(suite.T())
	mockDbusConnector := dbusMocks.NewMockConnector(suite.T())

	// Plan phase - no jobs but active shutdown timer
	mockDbusConnector.On("ListJobsContext", ctx).
		Return([]baseDbus.JobStatus{}, nil).
		Once()

	shutdownTimer := baseDbus.UnitStatus{
		Name:        "shutdown.timer",
		Description: "Shutdown Timer",
		LoadState:   "loaded",
		ActiveState: "active",
		SubState:    "waiting",
	}

	mockDbusConnector.On("ListUnitsContext", ctx).
		Return([]baseDbus.UnitStatus{shutdownTimer}, nil).
		Once()
	mockDbusConnector.On("Close").
		Return().
		Once()
	report := buildHostRebootOperator(suite, mockCmdExecutor, mockDbusConnector).Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"scheduled":true}`,
		"after":  `{"scheduled":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *HostRebootOperatorTestSuite) TestHostRebootOperatorRebootTimerDetection() {
	ctx := context.Background()

	mockCmdExecutor := utilsMocks.NewMockCommandExecutor(suite.T())
	mockDbusConnector := dbusMocks.NewMockConnector(suite.T())

	// Plan phase - no jobs but active reboot timer
	mockDbusConnector.On("ListJobsContext", ctx).
		Return([]baseDbus.JobStatus{}, nil).
		Once()

	rebootTimer := baseDbus.UnitStatus{
		Name:        "reboot.timer",
		Description: "Reboot Timer",
		LoadState:   "loaded",
		ActiveState: "active",
		SubState:    "waiting",
	}

	mockDbusConnector.On("ListUnitsContext", ctx).
		Return([]baseDbus.UnitStatus{rebootTimer}, nil).
		Once()
	mockDbusConnector.On("Close").
		Return().
		Once()
	report := buildHostRebootOperator(suite, mockCmdExecutor, mockDbusConnector).Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"scheduled":true}`,
		"after":  `{"scheduled":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *HostRebootOperatorTestSuite) TestHostRebootOperatorActiveShutdownProcessDetection() {
	ctx := context.Background()

	mockCmdExecutor := utilsMocks.NewMockCommandExecutor(suite.T())
	mockDbusConnector := dbusMocks.NewMockConnector(suite.T())

	// Plan phase - no jobs or timers but active shutdown process
	mockDbusConnector.On("ListJobsContext", ctx).
		Return([]baseDbus.JobStatus{}, nil).
		Once()

	mockDbusConnector.On("ListUnitsContext", ctx).
		Return([]baseDbus.UnitStatus{}, nil).
		Once()
	// Check for active shutdown processes - find shutdown process
	mockCmdExecutor.On("CombinedOutputContext", ctx, "pgrep", "-f", "shutdown").
		Return([]byte("1234"), nil).
		Once()
	mockDbusConnector.On("Close").
		Return().
		Once()

	report := buildHostRebootOperator(suite, mockCmdExecutor, mockDbusConnector).Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"scheduled":true}`,
		"after":  `{"scheduled":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *HostRebootOperatorTestSuite) TestHostRebootOperatorActiveSystemdShutdownProcessDetection() {
	ctx := context.Background()

	mockCmdExecutor := utilsMocks.NewMockCommandExecutor(suite.T())
	mockDbusConnector := dbusMocks.NewMockConnector(suite.T())

	// Plan phase - no jobs or timers, no shutdown process, but systemd-shutdown process
	mockDbusConnector.On("ListJobsContext", ctx).
		Return([]baseDbus.JobStatus{}, nil).
		Once()

	mockDbusConnector.On("ListUnitsContext", ctx).
		Return([]baseDbus.UnitStatus{}, nil).
		Once()
	// Check for active shutdown processes - no shutdown process
	mockCmdExecutor.On("CombinedOutputContext", ctx, "pgrep", "-f", "shutdown").
		Return([]byte(""), errors.New("no process found")).
		Once()
	// Find systemd-shutdown process
	mockCmdExecutor.On("CombinedOutputContext", ctx, "pgrep", "-f", "systemd-shutdown").
		Return([]byte("5678"), nil).
		Once()
	mockDbusConnector.On("Close").
		Return().
		Once()

	report := buildHostRebootOperator(suite, mockCmdExecutor, mockDbusConnector).Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"scheduled":true}`,
		"after":  `{"scheduled":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *HostRebootOperatorTestSuite) TestHostRebootOperatorScheduledFileDetection() {
	ctx := context.Background()

	mockCmdExecutor := utilsMocks.NewMockCommandExecutor(suite.T())
	mockDbusConnector := dbusMocks.NewMockConnector(suite.T())

	// Plan phase - no jobs, timers, or processes, but scheduled file exists
	mockDbusConnector.On("ListJobsContext", ctx).
		Return([]baseDbus.JobStatus{}, nil).
		Once()

	mockDbusConnector.On("ListUnitsContext", ctx).
		Return([]baseDbus.UnitStatus{}, nil).
		Once()
	// Check for active shutdown processes - none found
	mockCmdExecutor.On("CombinedOutputContext", ctx, "pgrep", "-f", "shutdown").
		Return([]byte(""), errors.New("no process found")).
		Once()
	mockCmdExecutor.On("CombinedOutputContext", ctx, "pgrep", "-f", "systemd-shutdown").
		Return([]byte(""), errors.New("no process found")).
		Once()
	// Find scheduled file
	mockCmdExecutor.On("CombinedOutputContext", ctx, "test", "-f", "/run/systemd/shutdown/scheduled").
		Return([]byte(""), nil).
		Once()
	mockDbusConnector.On("Close").
		Return().
		Once()

	report := buildHostRebootOperator(suite, mockCmdExecutor, mockDbusConnector).Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"scheduled":true}`,
		"after":  `{"scheduled":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}
