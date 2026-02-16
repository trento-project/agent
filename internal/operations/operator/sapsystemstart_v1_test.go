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

type SAPSystemStartOperatorTestSuite struct {
	suite.Suite
	mockSapcontrol *mocks.MockWebService
}

func TestSAPSystemStartOperator(t *testing.T) {
	suite.Run(t, new(SAPSystemStartOperatorTestSuite))
}

func (suite *SAPSystemStartOperatorTestSuite) SetupTest() {
	suite.mockSapcontrol = mocks.NewMockWebService(suite.T())
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartInstanceNumberMissing() {
	ctx := context.Background()

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.Arguments{},
		"test-op",
		operator.Options[operator.SAPSystemStart]{},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("argument instance_number not provided, could not use the operator", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartInstanceNumberInvalid() {
	ctx := context.Background()

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.Arguments{
			"instance_number": 0,
		},
		"test-op",
		operator.Options[operator.SAPSystemStart]{},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("could not parse instance_number argument as string, argument provided: 0", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartTimeoutInvalid() {
	ctx := context.Background()

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.Arguments{
			"instance_number": "00",
			"timeout":         "value",
		},
		"test-op",
		operator.Options[operator.SAPSystemStart]{},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("could not parse timeout argument as a number, argument provided: value", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartInstanceTypeInvalid() {
	ctx := context.Background()

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.Arguments{
			"instance_number": "00",
			"instance_type":   0,
		},
		"test-op",
		operator.Options[operator.SAPSystemStart]{},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("could not parse instance_type argument as a string, argument provided: 0", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartInstanceTypeUnknown() {
	ctx := context.Background()

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.Arguments{
			"instance_number": "00",
			"instance_type":   "unknown",
		},
		"test-op",
		operator.Options[operator.SAPSystemStart]{},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("invalid instance_type value: unknown", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartPlanError() {
	ctx := context.Background()

	suite.mockSapcontrol.
		On("GetSystemInstanceListContext", ctx, mock.Anything).
		Return(nil, errors.New("error getting instances")).
		Once()

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.Arguments{
			"instance_number": "00",
			"timeout":         300.0,
		},
		"test-op",
		operator.Options[operator.SAPSystemStart]{
			OperatorOptions: []operator.Option[operator.SAPSystemStart]{
				operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("error checking system state: error getting instance list: error getting instances", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartCommitAlreadyStarted() {
	ctx := context.Background()

	suite.mockSapcontrol.
		On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
		Return(&sapcontrolapi.GetSystemInstanceListResponse{
			Instances: []*sapcontrolapi.SAPInstance{
				{
					Dispstatus: sapcontrolapi.STATECOLOR_GREEN,
				},
				{
					Dispstatus: sapcontrolapi.STATECOLOR_GREEN,
				},
			},
		}, nil).
		Once()

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.Arguments{
			"instance_number": "00",
		},
		"test-op",
		operator.Options[operator.SAPSystemStart]{
			OperatorOptions: []operator.Option[operator.SAPSystemStart]{
				operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapSystemStartOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"started":true}`,
		"after":  `{"started":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartCommitAlreadyStartedFiltered() {
	cases := []struct {
		instanceType string
		features     string
	}{
		{
			instanceType: "abap",
			features:     "ABAP|GATEWAY|ICMAN|IGS",
		},
		{
			instanceType: "j2ee",
			features:     "J2EE|IGS",
		},
		{
			instanceType: "scs",
			features:     "MESSAGESERVER|ENQUE",
		},
		{
			instanceType: "enqrep",
			features:     "ENQREP",
		},
	}

	for _, tt := range cases {
		ctx := context.Background()
		suite.mockSapcontrol = mocks.NewMockWebService(suite.T())

		suite.mockSapcontrol.
			On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
			Return(&sapcontrolapi.GetSystemInstanceListResponse{
				Instances: []*sapcontrolapi.SAPInstance{
					{
						Features:   "Other",
						Dispstatus: sapcontrolapi.STATECOLOR_GRAY,
					},
					{
						Features:   tt.features,
						Dispstatus: sapcontrolapi.STATECOLOR_GREEN,
					},
				},
			}, nil).
			Once()

		sapSystemStartOperator := operator.NewSAPSystemStart(
			operator.Arguments{
				"instance_number": "00",
				"instance_type":   tt.instanceType,
			},
			"test-op",
			operator.Options[operator.SAPSystemStart]{
				OperatorOptions: []operator.Option[operator.SAPSystemStart]{
					operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemSapcontrol(suite.mockSapcontrol)),
				},
			},
		)

		report := sapSystemStartOperator.Run(ctx)

		expectedDiff := map[string]any{
			"before": `{"started":true}`,
			"after":  `{"started":true}`,
		}

		suite.Nil(report.Error)
		suite.Equal(operator.PLAN, report.Success.LastPhase)
		suite.EqualValues(expectedDiff, report.Success.Diff)
	}
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartCommitStartingError() {
	ctx := context.Background()

	planGetInstances := suite.mockSapcontrol.
		On("GetSystemInstanceListContext", ctx, mock.Anything).
		Return(
			&sapcontrolapi.GetSystemInstanceListResponse{
				Instances: []*sapcontrolapi.SAPInstance{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GRAY,
					},
				},
			}, nil,
		).
		Once()

	startSystem := suite.mockSapcontrol.
		On("StartSystemContext", ctx, mock.Anything).
		Return(nil, errors.New("error starting")).
		NotBefore(planGetInstances)

	rollbackStopSystem := suite.mockSapcontrol.
		On("StopSystemContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(startSystem)

	suite.mockSapcontrol.
		On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrolapi.GetSystemInstanceListResponse{
				Instances: []*sapcontrolapi.SAPInstance{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GRAY,
					},
				},
			}, nil,
		).
		Once().
		NotBefore(rollbackStopSystem)

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.Arguments{
			"instance_number": "00",
		},
		"test-op",
		operator.Options[operator.SAPSystemStart]{
			OperatorOptions: []operator.Option[operator.SAPSystemStart]{
				operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.EqualValues("error starting system: error starting", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartVerifyError() {
	ctx := context.Background()

	planGetInstances := suite.mockSapcontrol.
		On("GetSystemInstanceListContext", ctx, mock.Anything).
		Return(
			&sapcontrolapi.GetSystemInstanceListResponse{
				Instances: []*sapcontrolapi.SAPInstance{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GRAY,
					},
				},
			}, nil,
		).
		Once()

	startSystem := suite.mockSapcontrol.
		On("StartSystemContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(planGetInstances)

	verifyGetInstances := suite.mockSapcontrol.
		On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
		Return(nil, errors.New("error getting instances in verify")).
		Once().
		NotBefore(startSystem)

	rollbackStopSystem := suite.mockSapcontrol.
		On("StopSystemContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(verifyGetInstances)

	suite.mockSapcontrol.
		On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrolapi.GetSystemInstanceListResponse{
				Instances: []*sapcontrolapi.SAPInstance{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GRAY,
					},
				},
			}, nil,
		).
		Once().
		NotBefore(rollbackStopSystem)

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.Arguments{
			"instance_number": "00",
		},
		"test-op",
		operator.Options[operator.SAPSystemStart]{
			OperatorOptions: []operator.Option[operator.SAPSystemStart]{
				operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.EqualValues("verify system started failed: error getting instance list: error getting instances in verify", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartVerifyTimeout() {
	ctx := context.Background()

	suite.mockSapcontrol.
		On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrolapi.GetSystemInstanceListResponse{
				Instances: []*sapcontrolapi.SAPInstance{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GRAY,
					},
				},
			}, nil,
		).
		Times(3).
		On("StartSystemContext", ctx, mock.Anything).
		Return(nil, nil).
		On("StopSystemContext", ctx, mock.Anything).
		Return(nil, nil)

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.Arguments{
			"instance_number": "00",
			"timeout":         0.0,
		},
		"test-op",
		operator.Options[operator.SAPSystemStart]{
			OperatorOptions: []operator.Option[operator.SAPSystemStart]{
				operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemSapcontrol(suite.mockSapcontrol)),
				operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemInterval(0 * time.Second)),
			},
		},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues(
		"rollback to stopped failed: error waiting until system is in desired state\n"+
			"verify system started failed: "+
			"error waiting until system is in desired state", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartRollbackStoppingError() {
	ctx := context.Background()

	suite.mockSapcontrol.
		On("GetSystemInstanceListContext", ctx, mock.Anything).
		Return(
			&sapcontrolapi.GetSystemInstanceListResponse{
				Instances: []*sapcontrolapi.SAPInstance{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GRAY,
						Features:   "ABAP|GATEWAY|ICMAN|IGS",
					},
				},
			}, nil,
		).
		Once().
		On("StartSystemContext", ctx, mock.Anything).
		Return(nil, errors.New("error starting")).
		On(
			"StopSystemContext",
			ctx,
			mock.MatchedBy(func(req *sapcontrolapi.StopSystem) bool {
				return *req.Options == sapcontrolapi.StartStopOptionSAPControlABAPINSTANCES
			}),
		).
		Return(nil, errors.New("error stopping"))

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.Arguments{
			"instance_number": "00",
			"instance_type":   "abap",
		},
		"test-op",
		operator.Options[operator.SAPSystemStart]{
			OperatorOptions: []operator.Option[operator.SAPSystemStart]{
				operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues("error stopping system: error stopping\nerror starting system: error starting", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartSuccess() {
	cases := []struct {
		instanceType string
		features     string
		options      sapcontrolapi.StartStopOption
	}{
		{
			instanceType: "all",
			features:     "OTHER",
			options:      sapcontrolapi.StartStopOptionSAPControlALLINSTANCES,
		},
		{
			instanceType: "abap",
			features:     "ABAP|GATEWAY|ICMAN|IGS",
			options:      sapcontrolapi.StartStopOptionSAPControlABAPINSTANCES,
		},
		{
			instanceType: "j2ee",
			features:     "J2EE|IGS",
			options:      sapcontrolapi.StartStopOptionSAPControlJ2EEINSTANCES,
		},
		{
			instanceType: "scs",
			features:     "MESSAGESERVER|ENQUE",
			options:      sapcontrolapi.StartStopOptionSAPControlSCSINSTANCES,
		},
		{
			instanceType: "enqrep",
			features:     "ENQREP",
			options:      sapcontrolapi.StartStopOptionSAPControlENQREPINSTANCES,
		},
	}

	for _, tt := range cases {
		ctx := context.Background()
		suite.mockSapcontrol = mocks.NewMockWebService(suite.T())
		timeout := 60.0

		planGetInstances := suite.mockSapcontrol.
			On("GetSystemInstanceListContext", ctx, mock.Anything).
			Return(&sapcontrolapi.GetSystemInstanceListResponse{
				Instances: []*sapcontrolapi.SAPInstance{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GRAY,
						Features:   tt.features,
					},
				},
			}, nil).
			Once()

		suite.mockSapcontrol.
			On(
				"StartSystemContext",
				ctx,
				mock.MatchedBy(func(req *sapcontrolapi.StartSystem) bool {
					if *req.Options == tt.options && req.Waittimeout == int32(timeout) {
						return true
					}
					return false
				}),
			).
			Return(nil, nil).
			NotBefore(planGetInstances).
			On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
			Return(&sapcontrolapi.GetSystemInstanceListResponse{
				Instances: []*sapcontrolapi.SAPInstance{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GREEN,
						Features:   tt.features,
					},
				},
			}, nil).
			Once().
			NotBefore(planGetInstances)

		sapSystemStartOperator := operator.NewSAPSystemStart(
			operator.Arguments{
				"instance_number": "00",
				"instance_type":   tt.instanceType,
				"timeout":         timeout,
			},
			"test-op",
			operator.Options[operator.SAPSystemStart]{
				OperatorOptions: []operator.Option[operator.SAPSystemStart]{
					operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemSapcontrol(suite.mockSapcontrol)),
				},
			},
		)

		report := sapSystemStartOperator.Run(ctx)

		expectedDiff := map[string]any{
			"before": `{"started":false}`,
			"after":  `{"started":true}`,
		}

		suite.Nil(report.Error)
		suite.Equal(operator.VERIFY, report.Success.LastPhase)
		suite.EqualValues(expectedDiff, report.Success.Diff)
	}
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartSuccessMultipleQueries() {
	ctx := context.Background()

	planGetInstances := suite.mockSapcontrol.
		On("GetSystemInstanceListContext", ctx, mock.Anything).
		Return(
			&sapcontrolapi.GetSystemInstanceListResponse{
				Instances: []*sapcontrolapi.SAPInstance{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GRAY,
					},
				},
			}, nil,
		).
		Once()

	suite.mockSapcontrol.
		On("StartSystemContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(planGetInstances)

	suite.mockSapcontrol.
		On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrolapi.GetSystemInstanceListResponse{
				Instances: []*sapcontrolapi.SAPInstance{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GRAY,
					},
				},
			}, nil,
		).
		Times(3).
		NotBefore(planGetInstances).
		On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrolapi.GetSystemInstanceListResponse{
				Instances: []*sapcontrolapi.SAPInstance{
					{
						Dispstatus: sapcontrolapi.STATECOLOR_GREEN,
					},
				},
			}, nil,
		).
		Once()

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.Arguments{
			"instance_number": "00",
			"timeout":         5.0,
		},
		"test-op",
		operator.Options[operator.SAPSystemStart]{
			OperatorOptions: []operator.Option[operator.SAPSystemStart]{
				operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemSapcontrol(suite.mockSapcontrol)),
				operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemInterval(0 * time.Second)),
			},
		},
	)

	report := sapSystemStartOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"started":false}`,
		"after":  `{"started":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}
