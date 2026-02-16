package operator_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/sapsystem/sapcontrolapi"
	"github.com/trento-project/agent/internal/core/sapsystem/sapcontrolapi/mocks"
	"github.com/trento-project/agent/internal/operations/operator"
)

type SAPInstanceStopOperatorTestSuite struct {
	suite.Suite
	mockSapcontrol *mocks.MockWebService
}

func TestSAPInstanceStopOperator(t *testing.T) {
	suite.Run(t, new(SAPInstanceStopOperatorTestSuite))
}

func (suite *SAPInstanceStopOperatorTestSuite) SetupTest() {
	suite.mockSapcontrol = mocks.NewMockWebService(suite.T())
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopInstanceNumberMissing() {
	ctx := context.Background()

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.Arguments{},
		"test-op",
		operator.Options[operator.SAPInstanceStop]{},
	)

	report := sapInstanceStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("argument instance_number not provided, could not use the operator", report.Error.Message)
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopInstanceNumberInvalid() {
	ctx := context.Background()

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.Arguments{
			"instance_number": 0,
		},
		"test-op",
		operator.Options[operator.SAPInstanceStop]{},
	)

	report := sapInstanceStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("could not parse instance_number argument as string, argument provided: 0", report.Error.Message)
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopTimeoutInvalid() {
	ctx := context.Background()

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.Arguments{
			"instance_number": "00",
			"timeout":         "value",
		},
		"test-op",
		operator.Options[operator.SAPInstanceStop]{},
	)

	report := sapInstanceStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("could not parse timeout argument as a number, argument provided: value", report.Error.Message)
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopPlanError() {
	ctx := context.Background()

	suite.mockSapcontrol.
		On("GetProcessListContext", ctx, mock.Anything).
		Return(nil, errors.New("error getting processes")).
		Once()

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.Arguments{
			"instance_number": "00",
			"timeout":         300.0,
		},
		"test-op",
		operator.Options[operator.SAPInstanceStop]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStop]{
				operator.Option[operator.SAPInstanceStop](operator.WithCustomStopSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapInstanceStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("error checking processes state: error getting instance process list: error getting processes", report.Error.Message)
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopCommitAlreadyStopped() {
	ctx := context.Background()

	suite.mockSapcontrol.
		On("GetProcessListContext", mock.Anything, mock.Anything).
		Return(&sapcontrolapi.GetProcessListResponse{
			Processes: []*sapcontrolapi.OSProcess{
				{
					Dispstatus: sapcontrolapi.STATECOLOR_GRAY,
				},
				{
					Dispstatus: sapcontrolapi.STATECOLOR_GRAY,
				},
			},
		}, nil).
		Once()

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.Arguments{
			"instance_number": "00",
		},
		"test-op",
		operator.Options[operator.SAPInstanceStop]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStop]{
				operator.Option[operator.SAPInstanceStop](operator.WithCustomStopSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapInstanceStopOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"stopped":true}`,
		"after":  `{"stopped":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopCommitStoppingError() {
	ctx := context.Background()

	planGetProcesses := suite.mockSapcontrol.
		On("GetProcessListContext", ctx, mock.Anything).
		Return(
			&sapcontrolapi.GetProcessListResponse{
				Processes: []*sapcontrolapi.OSProcess{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GREEN,
					},
				},
			}, nil,
		).
		Once()

	suite.mockSapcontrol.
		On("StopContext", ctx, mock.Anything).
		Return(nil, errors.New("error stopping")).
		NotBefore(planGetProcesses).
		On("StartContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(planGetProcesses).
		On("GetProcessListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrolapi.GetProcessListResponse{
				Processes: []*sapcontrolapi.OSProcess{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GREEN,
					},
				},
			}, nil,
		).
		NotBefore(planGetProcesses).
		Once()

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.Arguments{
			"instance_number": "00",
		},
		"test-op",
		operator.Options[operator.SAPInstanceStop]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStop]{
				operator.Option[operator.SAPInstanceStop](operator.WithCustomStopSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapInstanceStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.EqualValues("error stopping instance: error stopping", report.Error.Message)
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopVerifyError() {
	ctx := context.Background()

	planGetProcesses := suite.mockSapcontrol.
		On("GetProcessListContext", ctx, mock.Anything).
		Return(
			&sapcontrolapi.GetProcessListResponse{
				Processes: []*sapcontrolapi.OSProcess{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GREEN,
					},
				},
			}, nil,
		).
		Once()

	verifyGetProcess := suite.mockSapcontrol.
		On("GetProcessListContext", mock.Anything, mock.Anything).
		Return(nil, errors.New("error getting processes in verify")).
		Once().
		NotBefore(planGetProcesses)

	suite.mockSapcontrol.
		On("StopContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(planGetProcesses).
		On("StartContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(verifyGetProcess).
		On("GetProcessListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrolapi.GetProcessListResponse{
				Processes: []*sapcontrolapi.OSProcess{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GREEN,
					},
				},
			}, nil,
		).
		Once().
		NotBefore(verifyGetProcess)

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.Arguments{
			"instance_number": "00",
		},
		"test-op",
		operator.Options[operator.SAPInstanceStop]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStop]{
				operator.Option[operator.SAPInstanceStop](operator.WithCustomStopSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapInstanceStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.EqualValues("verify instance stopped failed: error getting instance process list: error getting processes in verify", report.Error.Message)
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopVerifyTimeout() {
	ctx := context.Background()

	suite.mockSapcontrol.
		On("GetProcessListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrolapi.GetProcessListResponse{
				Processes: []*sapcontrolapi.OSProcess{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GREEN,
					},
				},
			}, nil,
		).
		Times(3).
		On("StopContext", ctx, mock.Anything).
		Return(nil, nil).
		On("StartContext", ctx, mock.Anything).
		Return(nil, nil)

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.Arguments{
			"instance_number": "00",
			"timeout":         0.0,
		},
		"test-op",
		operator.Options[operator.SAPInstanceStop]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStop]{
				operator.Option[operator.SAPInstanceStop](operator.WithCustomStopSapcontrol(suite.mockSapcontrol)),
				operator.Option[operator.SAPInstanceStop](operator.WithCustomStopInterval(0 * time.Second)),
			},
		},
	)

	report := sapInstanceStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues(
		"rollback to started failed: error waiting until instance is in desired state\n"+
			"verify instance stopped failed: "+
			"error waiting until instance is in desired state", report.Error.Message)
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopRollbackStartingError() {
	ctx := context.Background()

	suite.mockSapcontrol.
		On("GetProcessListContext", ctx, mock.Anything).
		Return(
			&sapcontrolapi.GetProcessListResponse{
				Processes: []*sapcontrolapi.OSProcess{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GREEN,
					},
				},
			}, nil,
		).
		Once().
		On("StopContext", ctx, mock.Anything).
		Return(nil, errors.New("error starting")).
		On("StartContext", ctx, mock.Anything).
		Return(nil, errors.New("error starting"))

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.Arguments{
			"instance_number": "00",
		},
		"test-op",
		operator.Options[operator.SAPInstanceStop]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStop]{
				operator.Option[operator.SAPInstanceStop](operator.WithCustomStopSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapInstanceStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues("error starting instance: error starting\nerror stopping instance: error starting", report.Error.Message)
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopSuccess() {
	ctx := context.Background()

	planGetProcesses := suite.mockSapcontrol.
		On("GetProcessListContext", ctx, mock.Anything).
		Return(&sapcontrolapi.GetProcessListResponse{
			Processes: []*sapcontrolapi.OSProcess{
				{
					Dispstatus: sapcontrolapi.STATECOLOR_GREEN,
				},
			},
		}, nil).
		Once()

	suite.mockSapcontrol.
		On("StopContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(planGetProcesses).
		On("GetProcessListContext", mock.Anything, mock.Anything).
		Return(&sapcontrolapi.GetProcessListResponse{
			Processes: []*sapcontrolapi.OSProcess{
				{
					Dispstatus: sapcontrolapi.STATECOLOR_GRAY,
				},
			},
		}, nil).
		Once().
		NotBefore(planGetProcesses)

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.Arguments{
			"instance_number": "00",
		},
		"test-op",
		operator.Options[operator.SAPInstanceStop]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStop]{
				operator.Option[operator.SAPInstanceStop](operator.WithCustomStopSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapInstanceStopOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"stopped":false}`,
		"after":  `{"stopped":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *SAPInstanceStartOperatorTestSuite) TestSAPInstanceStopSuccessMultipleQueries() {
	ctx := context.Background()

	planGetProcesses := suite.mockSapcontrol.
		On("GetProcessListContext", ctx, mock.Anything).
		Return(
			&sapcontrolapi.GetProcessListResponse{
				Processes: []*sapcontrolapi.OSProcess{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GREEN,
					},
				},
			}, nil,
		).
		Once()

	suite.mockSapcontrol.
		On("StopContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(planGetProcesses)

	suite.mockSapcontrol.
		On("GetProcessListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrolapi.GetProcessListResponse{
				Processes: []*sapcontrolapi.OSProcess{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GREEN,
					},
				},
			}, nil,
		).
		Times(3).
		NotBefore(planGetProcesses).
		On("GetProcessListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrolapi.GetProcessListResponse{
				Processes: []*sapcontrolapi.OSProcess{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GRAY,
					},
				},
			}, nil,
		).
		Once()

	sapInstanceStartOperator := operator.NewSAPInstanceStop(
		operator.Arguments{
			"instance_number": "00",
			"timeout":         5.0,
		},
		"test-op",
		operator.Options[operator.SAPInstanceStop]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStop]{
				operator.Option[operator.SAPInstanceStop](operator.WithCustomStopSapcontrol(suite.mockSapcontrol)),
				operator.Option[operator.SAPInstanceStop](operator.WithCustomStopInterval(0 * time.Second)),
			},
		},
	)

	report := sapInstanceStartOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"stopped":false}`,
		"after":  `{"stopped":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}
