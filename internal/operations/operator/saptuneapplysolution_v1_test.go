package operator_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/saptune/mocks"
	"github.com/trento-project/agent/internal/operations/operator"
)

type SaptuneApplySolutionOperatorTestSuite struct {
	suite.Suite
	mockSaptuneClient *mocks.MockSaptune
}

func TestSaptuneApplySolutionOperator(t *testing.T) {
	suite.Run(t, new(SaptuneApplySolutionOperatorTestSuite))
}

func (suite *SaptuneApplySolutionOperatorTestSuite) SetupTest() {
	suite.mockSaptuneClient = mocks.NewMockSaptune(suite.T())
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionPlanErrorParsingArguments() {
	ctx := context.Background()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.Arguments{
			"foo": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneApplySolution]{},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("argument solution not provided, could not use the operator", report.Error.Message)
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionPlanErrorEmptySolutionRequested() {
	ctx := context.Background()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.Arguments{
			"solution": "",
		},
		"test-op",
		operator.Options[operator.SaptuneApplySolution]{},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("solution argument is empty", report.Error.Message)
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionPlanErrorVersionCheck() {
	ctx := context.Background()

	suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(errors.New("saptune version not supported")).
		Once()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClientApply(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("saptune version not supported", report.Error.Message)
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionPlanErrorGettingSolution() {
	ctx := context.Background()

	checkSaptuneVersionCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", errors.New("failed to determine initially applied solution")).
		NotBefore(checkSaptuneVersionCall).
		Once()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClientApply(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("failed to determine initially applied solution", report.Error.Message)
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionCommitErrorAnotherSolutionAlreadyAppliedWithSuccessfulRollback() {
	ctx := context.Background()

	checkSaptuneVersionCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("HANA", nil).
		NotBefore(checkSaptuneVersionCall).
		Once()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.Arguments{
			"solution": "S4HANA-DBSERVER",
		},
		"test-op",
		operator.Options[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClientApply(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.EqualValues("cannot apply solution S4HANA-DBSERVER because another solution HANA is already applied", report.Error.Message)
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionCommitErrorApplyingSolutionWithSuccessfulRollback() {
	ctx := context.Background()

	checkVersionSupportCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	getAppliedSolutionCall := suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", nil).
		NotBefore(checkVersionSupportCall).
		Once()

	applySolutionCall := suite.mockSaptuneClient.On(
		"ApplySolution",
		ctx,
		"HANA",
	).Return(errors.New("failed to apply solution")).
		NotBefore(checkVersionSupportCall).
		NotBefore(getAppliedSolutionCall).
		Once()

	suite.mockSaptuneClient.On(
		"RevertSolution",
		ctx,
		"HANA",
	).Return(nil).
		NotBefore(checkVersionSupportCall).
		NotBefore(getAppliedSolutionCall).
		NotBefore(applySolutionCall).
		Once()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClientApply(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.EqualValues("failed to apply solution", report.Error.Message)
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionCommitErrorApplyingSolutionWithFailingRollback() {
	ctx := context.Background()

	checkVersionSupportCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	getAppliedSolutionCall := suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", nil).
		NotBefore(checkVersionSupportCall).
		Once()

	applySolutionCall := suite.mockSaptuneClient.On(
		"ApplySolution",
		ctx,
		"HANA",
	).Return(errors.New("failed to apply solution")).
		NotBefore(checkVersionSupportCall).
		NotBefore(getAppliedSolutionCall).
		Once()

	suite.mockSaptuneClient.On(
		"RevertSolution",
		ctx,
		"HANA",
	).Return(errors.New("failed to revert solution")).
		NotBefore(checkVersionSupportCall).
		NotBefore(getAppliedSolutionCall).
		NotBefore(applySolutionCall).
		Once()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClientApply(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.Contains(report.Error.Message, "failed to apply solution")
	suite.Contains(report.Error.Message, "failed to revert solution")
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionVerifyErrorDeterminingAppliedSolutionWithSuccessfulRollback() {
	ctx := context.Background()

	checkVersionSupportCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	getAppliedSolutionCall := suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", nil).
		NotBefore(checkVersionSupportCall).
		Once()

	applySolutionCall := suite.mockSaptuneClient.On(
		"ApplySolution",
		ctx,
		"HANA",
	).Return(nil).
		NotBefore(checkVersionSupportCall).
		NotBefore(getAppliedSolutionCall).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", errors.New("failed to determine applied solution during verify")).
		Once()

	suite.mockSaptuneClient.On(
		"RevertSolution",
		ctx,
		"HANA",
	).Return(nil).
		NotBefore(checkVersionSupportCall).
		NotBefore(getAppliedSolutionCall).
		NotBefore(applySolutionCall).
		Once()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClientApply(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.EqualValues("failed to determine applied solution during verify", report.Error.Message)
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionVerifyErrorDeterminingAppliedSolutionWithFailingRollback() {
	ctx := context.Background()

	checkVersionSupportCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	getAppliedSolutionCall := suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", nil).
		NotBefore(checkVersionSupportCall).
		Once()

	applySolutionCall := suite.mockSaptuneClient.On(
		"ApplySolution",
		ctx,
		"HANA",
	).Return(nil).
		NotBefore(checkVersionSupportCall).
		NotBefore(getAppliedSolutionCall).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", errors.New("failed to determine applied solution during verify")).
		Once()

	suite.mockSaptuneClient.On(
		"RevertSolution",
		ctx,
		"HANA",
	).Return(errors.New("failed to revert solution")).
		NotBefore(checkVersionSupportCall).
		NotBefore(getAppliedSolutionCall).
		NotBefore(applySolutionCall).
		Once()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClientApply(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.Contains(report.Error.Message, "failed to determine applied solution during verify")
	suite.Contains(report.Error.Message, "failed to revert solution")
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionVerifyErrorAppliedSolutionDiffersFromRequestedWithSuccessfulRollback() {
	ctx := context.Background()

	checkVersionSupportCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	firstGetAppliedSolutionCall := suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", nil).
		NotBefore(checkVersionSupportCall).
		Once()

	applySolutionCall := suite.mockSaptuneClient.On(
		"ApplySolution",
		ctx,
		"HANA",
	).Return(nil).
		NotBefore(checkVersionSupportCall).
		NotBefore(firstGetAppliedSolutionCall).
		Once()

	getAppliedSolutionCall := suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("S4HANA-DBSERVER", nil).
		NotBefore(checkVersionSupportCall).
		NotBefore(firstGetAppliedSolutionCall).
		NotBefore(applySolutionCall).
		Once()

	suite.mockSaptuneClient.On(
		"RevertSolution",
		ctx,
		"HANA",
	).Return(nil).
		NotBefore(checkVersionSupportCall).
		NotBefore(firstGetAppliedSolutionCall).
		NotBefore(applySolutionCall).
		NotBefore(getAppliedSolutionCall).
		Once()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClientApply(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.EqualValues("verify saptune apply failing, the solution HANA was not applied in commit phase", report.Error.Message)
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionErrorDetectedAppliedSolutionDiffersFromRequestedWithFailingRollback() {
	ctx := context.Background()

	checkVersionSupportCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	firstGetAppliedSolutionCall := suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", nil).
		NotBefore(checkVersionSupportCall).
		Once()

	applySolutionCall := suite.mockSaptuneClient.On(
		"ApplySolution",
		ctx,
		"HANA",
	).Return(nil).
		NotBefore(checkVersionSupportCall).
		NotBefore(firstGetAppliedSolutionCall).
		Once()

	getAppliedSolutionCall := suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("S4HANA-DBSERVER", nil).
		NotBefore(checkVersionSupportCall).
		NotBefore(firstGetAppliedSolutionCall).
		NotBefore(applySolutionCall).
		Once()

	suite.mockSaptuneClient.On(
		"RevertSolution",
		ctx,
		"HANA",
	).Return(errors.New("failed to revert solution")).
		NotBefore(checkVersionSupportCall).
		NotBefore(firstGetAppliedSolutionCall).
		NotBefore(applySolutionCall).
		NotBefore(getAppliedSolutionCall).
		Once()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClientApply(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.Contains(report.Error.Message, "verify saptune apply failing, the solution HANA was not applied in commit phase")
	suite.Contains(report.Error.Message, "failed to revert solution")
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionSuccess() {
	ctx := context.Background()

	checkSaptuneVersionCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	solutionAppliedCall := suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", nil).
		NotBefore(checkSaptuneVersionCall).
		Once()

	solutionApplyCall := suite.mockSaptuneClient.On(
		"ApplySolution",
		ctx,
		"HANA",
	).Return(nil).
		NotBefore(solutionAppliedCall).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("HANA", nil).
		NotBefore(solutionApplyCall).
		Once()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClientApply(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"solution":""}`,
		"after":  `{"solution":"HANA"}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *SaptuneApplySolutionOperatorTestSuite) TestSaptuneApplySolutionSuccessReapplyingAlreadyApplied() {
	ctx := context.Background()

	checkSaptuneVersionCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("HANA", nil).
		NotBefore(checkSaptuneVersionCall).
		Once()

	saptuneSolutionApplyOperator := operator.NewSaptuneApplySolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneApplySolution]{
			OperatorOptions: []operator.Option[operator.SaptuneApplySolution]{
				operator.Option[operator.SaptuneApplySolution](operator.WithSaptuneClientApply(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"solution":"HANA"}`,
		"after":  `{"solution":"HANA"}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}
