package operator_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/saptune/mocks"
	"github.com/trento-project/agent/internal/operations/operator"
)

type SaptuneChangeSolutionOperatorTestSuite struct {
	suite.Suite
	mockSaptuneClient *mocks.MockSaptune
}

func TestSaptuneChangeSolutionOperator(t *testing.T) {
	suite.Run(t, new(SaptuneChangeSolutionOperatorTestSuite))
}

func (suite *SaptuneChangeSolutionOperatorTestSuite) SetupTest() {
	suite.mockSaptuneClient = mocks.NewMockSaptune(suite.T())
}

func (suite *SaptuneChangeSolutionOperatorTestSuite) TestSaptuneChangeSolutionPlanErrorParsingArguments() {
	ctx := context.Background()

	saptuneSolutionChangeOperator := operator.NewSaptuneChangeSolution(
		operator.Arguments{
			"foo": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneChangeSolution]{},
	)

	report := saptuneSolutionChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("argument solution not provided, could not use the operator", report.Error.Message)
}

func (suite *SaptuneChangeSolutionOperatorTestSuite) TestSaptuneChangeSolutionPlanErrorEmptySolutionRequested() {
	ctx := context.Background()

	saptuneSolutionChangeOperator := operator.NewSaptuneChangeSolution(
		operator.Arguments{
			"solution": "",
		},
		"test-op",
		operator.Options[operator.SaptuneChangeSolution]{},
	)

	report := saptuneSolutionChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("solution argument is empty", report.Error.Message)
}

func (suite *SaptuneChangeSolutionOperatorTestSuite) TestSaptuneChangeSolutionPlanErrorVersionCheck() {
	ctx := context.Background()

	suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(errors.New("saptune version not supported")).
		Once()

	saptuneSolutionChangeOperator := operator.NewSaptuneChangeSolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneChangeSolution]{
			OperatorOptions: []operator.Option[operator.SaptuneChangeSolution]{
				operator.Option[operator.SaptuneChangeSolution](operator.WithSaptuneClientChange(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("saptune version not supported", report.Error.Message)
}

func (suite *SaptuneChangeSolutionOperatorTestSuite) TestSaptuneChangeSolutionPlanErrorGettingSolution() {
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

	saptuneSolutionApplyOperator := operator.NewSaptuneChangeSolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneChangeSolution]{
			OperatorOptions: []operator.Option[operator.SaptuneChangeSolution]{
				operator.Option[operator.SaptuneChangeSolution](operator.WithSaptuneClientChange(suite.mockSaptuneClient)),
			},
		},
	)

	report := saptuneSolutionApplyOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("failed to determine initially applied solution", report.Error.Message)
}

func (suite *SaptuneChangeSolutionOperatorTestSuite) TestSaptuneChangeSolutionCommitErrorNoPreviouslyAppliedSolution() {

	ctx := context.Background()

	checkSaptuneVersionCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", nil).
		NotBefore(checkSaptuneVersionCall).
		Once()

	saptuneSolutionChangeOperator := operator.NewSaptuneChangeSolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneChangeSolution]{
			OperatorOptions: []operator.Option[operator.SaptuneChangeSolution]{
				operator.Option[operator.SaptuneChangeSolution](operator.WithSaptuneClientChange(suite.mockSaptuneClient)),
			},
		},
	)
	report := saptuneSolutionChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.Contains(report.Error.Message, "cannot change solution to HANA because no solution is currently applied")

}

func (suite *SaptuneChangeSolutionOperatorTestSuite) TestSaptuneChangeSolutionCommitErrorChangeSolutionSuccessfulRollback() {
	ctx := context.Background()

	checkSaptuneVersionCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	getAppliedSolutionCall := suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("FOO", nil).
		NotBefore(checkSaptuneVersionCall).
		Once()

	changeSolutionCall := suite.mockSaptuneClient.On(
		"ChangeSolution",
		ctx,
		"HANA",
	).Return(errors.New("failed to change solution")).
		NotBefore(checkSaptuneVersionCall).
		NotBefore(getAppliedSolutionCall).
		Once()

	suite.mockSaptuneClient.On(
		"ChangeSolution",
		ctx,
		"FOO",
	).Return(nil).
		NotBefore(checkSaptuneVersionCall).
		NotBefore(getAppliedSolutionCall).
		NotBefore(changeSolutionCall).
		Once()

	saptuneSolutionChangeOperator := operator.NewSaptuneChangeSolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneChangeSolution]{
			OperatorOptions: []operator.Option[operator.SaptuneChangeSolution]{
				operator.Option[operator.SaptuneChangeSolution](operator.WithSaptuneClientChange(suite.mockSaptuneClient)),
			},
		},
	)
	report := saptuneSolutionChangeOperator.Run(ctx)
	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.Contains(report.Error.Message, "failed to change solution")

}

func (suite *SaptuneChangeSolutionOperatorTestSuite) TestSaptuneChangeSolutionCommitErrorChangeSolutionFailingRollback() {
	ctx := context.Background()

	checkSaptuneVersionCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	getAppliedSolutionCall := suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("FOO", nil).
		NotBefore(checkSaptuneVersionCall).
		Once()

	changeSolutionCall := suite.mockSaptuneClient.On(
		"ChangeSolution",
		ctx,
		"HANA",
	).Return(errors.New("failed to change solution")).
		NotBefore(checkSaptuneVersionCall).
		NotBefore(getAppliedSolutionCall).
		Once()

	suite.mockSaptuneClient.On(
		"ChangeSolution",
		ctx,
		"FOO",
	).Return(errors.New("failed to change back to initial solution")).
		NotBefore(checkSaptuneVersionCall).
		NotBefore(getAppliedSolutionCall).
		NotBefore(changeSolutionCall).
		Once()

	saptuneSolutionChangeOperator := operator.NewSaptuneChangeSolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneChangeSolution]{
			OperatorOptions: []operator.Option[operator.SaptuneChangeSolution]{
				operator.Option[operator.SaptuneChangeSolution](operator.WithSaptuneClientChange(suite.mockSaptuneClient)),
			},
		},
	)
	report := saptuneSolutionChangeOperator.Run(ctx)
	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.Contains(report.Error.Message, "failed to change solution")
	suite.Contains(report.Error.Message, "failed to change back to initial solution")
}

func (suite *SaptuneChangeSolutionOperatorTestSuite) TestSaptuneChangeSolutionVerifyErrorGetAppliedSolutionSuccessfulRollback() {
	ctx := context.Background()

	checkSaptuneVersionCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	getAppliedSolutionCall := suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("FOO", nil).
		NotBefore(checkSaptuneVersionCall).
		Once()

	initialChangeSolutionCall := suite.mockSaptuneClient.On(
		"ChangeSolution",
		ctx,
		"HANA",
	).Return(nil).
		NotBefore(checkSaptuneVersionCall).
		NotBefore(getAppliedSolutionCall).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", errors.New("failed to determine currently applied solution in verify")).
		NotBefore(checkSaptuneVersionCall).
		NotBefore(getAppliedSolutionCall).
		NotBefore(initialChangeSolutionCall).
		Once()

	suite.mockSaptuneClient.On(
		"ChangeSolution",
		ctx,
		"FOO",
	).Return(nil).
		NotBefore(checkSaptuneVersionCall).
		NotBefore(getAppliedSolutionCall).
		NotBefore(initialChangeSolutionCall).
		Once()

	saptuneSolutionChangeOperator := operator.NewSaptuneChangeSolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneChangeSolution]{
			OperatorOptions: []operator.Option[operator.SaptuneChangeSolution]{
				operator.Option[operator.SaptuneChangeSolution](operator.WithSaptuneClientChange(suite.mockSaptuneClient)),
			},
		},
	)
	report := saptuneSolutionChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.Contains(report.Error.Message, "failed to determine currently applied solution in verify")
}

func (suite *SaptuneChangeSolutionOperatorTestSuite) TestSaptuneChangeSolutionVerifyErrorGetAppliedSolutionFailingRollback() {
	ctx := context.Background()

	checkSaptuneVersionCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	getAppliedSolutionCall := suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("FOO", nil).
		NotBefore(checkSaptuneVersionCall).
		Once()

	initialChangeSolutionCall := suite.mockSaptuneClient.On(
		"ChangeSolution",
		ctx,
		"HANA",
	).Return(nil).
		NotBefore(checkSaptuneVersionCall).
		NotBefore(getAppliedSolutionCall).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("", errors.New("failed to determine currently applied solution in verify")).
		NotBefore(checkSaptuneVersionCall).
		NotBefore(getAppliedSolutionCall).
		NotBefore(initialChangeSolutionCall).
		Once()

	suite.mockSaptuneClient.On(
		"ChangeSolution",
		ctx,
		"FOO",
	).Return(errors.New("failed to change back to initial solution")).
		NotBefore(checkSaptuneVersionCall).
		NotBefore(getAppliedSolutionCall).
		NotBefore(initialChangeSolutionCall).
		Once()

	saptuneSolutionChangeOperator := operator.NewSaptuneChangeSolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneChangeSolution]{
			OperatorOptions: []operator.Option[operator.SaptuneChangeSolution]{
				operator.Option[operator.SaptuneChangeSolution](operator.WithSaptuneClientChange(suite.mockSaptuneClient)),
			},
		},
	)
	report := saptuneSolutionChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.Contains(report.Error.Message, "failed to determine currently applied solution in verify")
	suite.Contains(report.Error.Message, "failed to change back to initial solution")
}

func (suite *SaptuneChangeSolutionOperatorTestSuite) TestSaptuneChangeSolutionVerifyErrorNonMatchingSuccessfulRollback() {
	ctx := context.Background()

	checkSaptuneVersionCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	getAppliedSolutionCall := suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("FOO", nil).
		NotBefore(checkSaptuneVersionCall).
		Once()

	initialChangeSolutionCall := suite.mockSaptuneClient.On(
		"ChangeSolution",
		ctx,
		"HANA",
	).Return(nil).
		NotBefore(checkSaptuneVersionCall).
		NotBefore(getAppliedSolutionCall).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("NOT_HANA", nil).
		NotBefore(checkSaptuneVersionCall).
		NotBefore(getAppliedSolutionCall).
		NotBefore(initialChangeSolutionCall).
		Once()

	suite.mockSaptuneClient.On(
		"ChangeSolution",
		ctx,
		"FOO",
	).Return(nil).
		NotBefore(checkSaptuneVersionCall).
		NotBefore(getAppliedSolutionCall).
		NotBefore(initialChangeSolutionCall).
		Once()

	saptuneSolutionChangeOperator := operator.NewSaptuneChangeSolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneChangeSolution]{
			OperatorOptions: []operator.Option[operator.SaptuneChangeSolution]{
				operator.Option[operator.SaptuneChangeSolution](operator.WithSaptuneClientChange(suite.mockSaptuneClient)),
			},
		},
	)
	report := saptuneSolutionChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.Contains(report.Error.Message, "verify saptune apply failing, the solution HANA was not applied in commit phase")
}

func (suite *SaptuneChangeSolutionOperatorTestSuite) TestSaptuneChangeSolutionVerifyErrorNonMatchingFailingRollback() {
	ctx := context.Background()

	checkSaptuneVersionCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	getAppliedSolutionCall := suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("FOO", nil).
		NotBefore(checkSaptuneVersionCall).
		Once()

	initialChangeSolutionCall := suite.mockSaptuneClient.On(
		"ChangeSolution",
		ctx,
		"HANA",
	).Return(nil).
		NotBefore(checkSaptuneVersionCall).
		NotBefore(getAppliedSolutionCall).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("NOT_HANA", nil).
		NotBefore(checkSaptuneVersionCall).
		NotBefore(getAppliedSolutionCall).
		NotBefore(initialChangeSolutionCall).
		Once()

	suite.mockSaptuneClient.On(
		"ChangeSolution",
		ctx,
		"FOO",
	).Return(errors.New("failed to change back to initial solution")).
		NotBefore(checkSaptuneVersionCall).
		NotBefore(getAppliedSolutionCall).
		NotBefore(initialChangeSolutionCall).
		Once()

	saptuneSolutionChangeOperator := operator.NewSaptuneChangeSolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneChangeSolution]{
			OperatorOptions: []operator.Option[operator.SaptuneChangeSolution]{
				operator.Option[operator.SaptuneChangeSolution](operator.WithSaptuneClientChange(suite.mockSaptuneClient)),
			},
		},
	)
	report := saptuneSolutionChangeOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.Contains(report.Error.Message, "verify saptune apply failing, the solution HANA was not applied in commit phase")
	suite.Contains(report.Error.Message, "failed to change back to initial solution")
}

func (suite *SaptuneChangeSolutionOperatorTestSuite) TestSaptuneChangeSolutionSuccessNoChange() {
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

	saptuneSolutionChangeOperator := operator.NewSaptuneChangeSolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneChangeSolution]{
			OperatorOptions: []operator.Option[operator.SaptuneChangeSolution]{
				operator.Option[operator.SaptuneChangeSolution](operator.WithSaptuneClientChange(suite.mockSaptuneClient)),
			},
		},
	)
	report := saptuneSolutionChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"solution":"HANA"}`,
		"after":  `{"solution":"HANA"}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *SaptuneChangeSolutionOperatorTestSuite) TestSaptuneChangeSolutionSuccess() {
	ctx := context.Background()

	checkSaptuneVersionCall := suite.mockSaptuneClient.On(
		"CheckVersionSupport",
		ctx,
	).Return(nil).
		Once()

	initialGetAppliedSolutionCall := suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("FOO", nil).
		NotBefore(checkSaptuneVersionCall).
		Once()

	changeSolutionCall := suite.mockSaptuneClient.On(
		"ChangeSolution",
		ctx,
		"HANA",
	).Return(nil).
		NotBefore(checkSaptuneVersionCall).
		NotBefore(initialGetAppliedSolutionCall).
		Once()

	suite.mockSaptuneClient.On(
		"GetAppliedSolution",
		ctx,
	).Return("HANA", nil).
		NotBefore(checkSaptuneVersionCall).
		NotBefore(initialGetAppliedSolutionCall).
		NotBefore(changeSolutionCall).
		Once()

	saptuneSolutionChangeOperator := operator.NewSaptuneChangeSolution(
		operator.Arguments{
			"solution": "HANA",
		},
		"test-op",
		operator.Options[operator.SaptuneChangeSolution]{
			OperatorOptions: []operator.Option[operator.SaptuneChangeSolution]{
				operator.Option[operator.SaptuneChangeSolution](operator.WithSaptuneClientChange(suite.mockSaptuneClient)),
			},
		},
	)
	report := saptuneSolutionChangeOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"solution":"FOO"}`,
		"after":  `{"solution":"HANA"}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}
