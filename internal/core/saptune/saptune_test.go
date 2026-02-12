package saptune_test

import (
	"context"
	"errors"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/saptune"
	"github.com/trento-project/agent/pkg/utils/mocks"
	"github.com/trento-project/agent/test/helpers"
)

type SaptuneClientTestSuite struct {
	suite.Suite
	mockExecutor *mocks.MockCommandExecutor
	logger       *slog.Logger
}

func TestSaptuneClient(t *testing.T) {
	suite.Run(t, new(SaptuneClientTestSuite))
}

func (suite *SaptuneClientTestSuite) SetupTest() {
	suite.mockExecutor = mocks.NewMockCommandExecutor(suite.T())
	suite.logger = slog.Default()
}

func (suite *SaptuneClientTestSuite) TestGetVersion() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"rpm",
		"-q",
		"--qf",
		"%{VERSION}",
		"saptune",
	).Return(
		[]byte("3.1.0"),
		nil,
	)

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	version, err := saptuneClient.GetVersion(ctx)

	suite.NoError(err)
	suite.Equal("3.1.0", version)
}

func (suite *SaptuneClientTestSuite) TestGetVersionError() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"rpm",
		"-q",
		"--qf",
		"%{VERSION}",
		"saptune",
	).Return(
		[]byte("package saptune is not installed"),
		errors.New("exit status 1"),
	)

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	_, err := saptuneClient.GetVersion(ctx)

	suite.Error(err)
	suite.ErrorContains(err, "could not get the installed saptune version")
}

func (suite *SaptuneClientTestSuite) TestVersionCheckFailureBecauseUnableToDetectVersion() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"rpm",
		"-q",
		"--qf",
		"%{VERSION}",
		"saptune",
	).Return(
		[]byte("package saptune is not installed"),
		errors.New("exit status 1"),
	)

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	err := saptuneClient.CheckVersionSupport(ctx)

	suite.Error(err)
	suite.ErrorContains(err, "could not get the installed saptune version")
}

func (suite *SaptuneClientTestSuite) TestUnsupportedSaptuneVersionCheck() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"rpm",
		"-q",
		"--qf",
		"%{VERSION}",
		"saptune",
	).Return([]byte("3.0.2"), nil)

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	err := saptuneClient.CheckVersionSupport(ctx)

	suite.Error(err)
	suite.ErrorContains(err, "saptune version not supported")
}

func (suite *SaptuneClientTestSuite) TestSuccessfulSaptuneVersionCheck() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"rpm",
		"-q",
		"--qf",
		"%{VERSION}",
		"saptune",
	).Return([]byte("3.1.0"), nil)

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	err := saptuneClient.CheckVersionSupport(ctx)

	suite.NoError(err)
}

func (suite *SaptuneClientTestSuite) TestGettingAppliedSolutionFailure() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"saptune",
		"--format",
		"json",
		"solution",
		"applied",
	).Return(nil, errors.New("error calling saptune"))

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	appliedSolution, err := saptuneClient.GetAppliedSolution(ctx)

	suite.Error(err)
	suite.ErrorContains(err, "error executing saptune command: error calling saptune")
	suite.Empty(appliedSolution)
}

func (suite *SaptuneClientTestSuite) TestGettingNoSolutionApplied() {
	ctx := context.Background()

	noSolutionApplied := helpers.ReadFixture("saptune/applied_no_solution.json")

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"saptune",
		"--format",
		"json",
		"solution",
		"applied",
	).Return(noSolutionApplied, nil)

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	appliedSolution, err := saptuneClient.GetAppliedSolution(ctx)

	suite.NoError(err)
	suite.Empty(appliedSolution)
}

func (suite *SaptuneClientTestSuite) TestGettingAppliedSolution() {
	ctx := context.Background()

	hanaSolutionApplied := helpers.ReadFixture("saptune/applied_hana_solution.json")

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"saptune",
		"--format",
		"json",
		"solution",
		"applied",
	).Return(hanaSolutionApplied, nil)

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	appliedSolution, err := saptuneClient.GetAppliedSolution(ctx)

	suite.NoError(err)
	suite.Equal("HANA", appliedSolution)
}

func (suite *SaptuneClientTestSuite) TestApplySolutionFailureBecauseCommandFails() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"saptune",
		"solution",
		"apply",
		"HANA",
	).Return(nil, errors.New("error calling saptune"))

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	err := saptuneClient.ApplySolution(ctx, "HANA")

	suite.Error(err)
	suite.ErrorContains(err, `error executing saptune command: error calling saptune`)
}

func (suite *SaptuneClientTestSuite) TestApplySolutionFailureBecauseAnAlreadyAppliedSolution() {
	ctx := context.Background()

	alreadyAppliedSolution := helpers.ReadFixture("saptune/apply_already_applied_solution.output")

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"saptune",
		"solution",
		"apply",
		"HANA",
	).Return(alreadyAppliedSolution, errors.New("exit status 1"))

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	err := saptuneClient.ApplySolution(ctx, "HANA")

	suite.Error(err)
	suite.ErrorContains(err, `error executing saptune command: exit status 1`)
}

func (suite *SaptuneClientTestSuite) TestApplySolutionSuccess() {
	ctx := context.Background()

	applySolutionSuccess := helpers.ReadFixture("saptune/apply_solution_success.output")

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"saptune",
		"solution",
		"apply",
		"HANA",
	).Return(applySolutionSuccess, nil)

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	err := saptuneClient.ApplySolution(ctx, "HANA")

	suite.NoError(err)
}

func (suite *SaptuneClientTestSuite) TestRevertSolutionFailureBecauseCommandFails() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"saptune",
		"solution",
		"revert",
		"HANA",
	).Return(nil, errors.New("error calling saptune"))

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	err := saptuneClient.RevertSolution(ctx, "HANA")

	suite.Error(err)
	suite.ErrorContains(err, `error executing saptune command: error calling saptune`)
}

func (suite *SaptuneClientTestSuite) TestRevertSolutionSuccess() {
	scenarios := []struct {
		name          string
		commandOutput []byte
	}{
		{
			name:          "reverting applied solution",
			commandOutput: helpers.ReadFixture("saptune/revert_solution_success.output"),
		},
		{
			name:          "reverting not applied solution",
			commandOutput: helpers.ReadFixture("saptune/revert_not_applied_solution.output"),
		},
	}

	for _, scenario := range scenarios {
		ctx := context.Background()

		suite.mockExecutor.On(
			"CombinedOutputContext",
			ctx,
			"saptune",
			"solution",
			"revert",
			"HANA",
		).Return(scenario.commandOutput, nil)

		saptuneClient := saptune.NewSaptuneClient(
			suite.mockExecutor,
			suite.logger,
		)
		err := saptuneClient.RevertSolution(ctx, "HANA")

		suite.NoError(err)
	}
}

func (suite *SaptuneClientTestSuite) TestChangeSolutionFailureBecauseCommandFails() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"saptune",
		"solution",
		"change",
		"--force",
		"HANA",
	).Return(nil, errors.New("error calling saptune"))

	saptuneClient := saptune.NewSaptuneClient(
		suite.mockExecutor,
		suite.logger,
	)
	err := saptuneClient.ChangeSolution(ctx, "HANA")

	suite.Error(err)
	suite.ErrorContains(err, `error executing saptune command: error calling saptune`)
}

func (suite *SaptuneClientTestSuite) TestChangeSolutionSucceess() {
	ctx := context.Background()

	scenarios := []struct {
		name          string
		commandOutput []byte
	}{
		{
			name:          "change solution",
			commandOutput: helpers.ReadFixture("saptune/change_solution_success.output"),
		},
		{
			name:          "change to already applied solution",
			commandOutput: helpers.ReadFixture("saptune/change_solution_to_same_success.output"),
		},
	}
	for _, scenario := range scenarios {
		suite.mockExecutor.On(
			"CombinedOutputContext",
			ctx,
			"saptune",
			"solution",
			"change",
			"--force",
			"HANA",
		).Return(scenario.commandOutput, nil)

		saptuneClient := saptune.NewSaptuneClient(
			suite.mockExecutor,
			suite.logger,
		)
		err := saptuneClient.ChangeSolution(ctx, "HANA")

		suite.NoError(err)
	}
}

func (suite *SaptuneClientTestSuite) TestCheck() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"saptune",
		"--format",
		"json",
		"check",
	).Return([]byte("check output"), nil)

	saptuneClient := saptune.NewSaptuneClient(suite.mockExecutor, suite.logger)
	checkOutput, err := saptuneClient.Check(ctx)

	suite.NoError(err)
	suite.Equal([]byte("check output"), checkOutput)
}

func (suite *SaptuneClientTestSuite) TestCheckError() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"saptune",
		"--format",
		"json",
		"check",
	).Return(nil, errors.New("check error"))

	saptuneClient := saptune.NewSaptuneClient(suite.mockExecutor, suite.logger)
	_, err := saptuneClient.Check(ctx)

	suite.Error(err)
	suite.ErrorContains(err, "error executing saptune command: check error")
}

func (suite *SaptuneClientTestSuite) TestListSolution() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"saptune",
		"--format",
		"json",
		"solution",
		"list",
	).Return([]byte("list solution output"), nil)

	saptuneClient := saptune.NewSaptuneClient(suite.mockExecutor, suite.logger)
	listSolutionOutput, err := saptuneClient.ListSolution(ctx)

	suite.NoError(err)
	suite.Equal([]byte("list solution output"), listSolutionOutput)
}

func (suite *SaptuneClientTestSuite) TestListSolutionError() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"saptune",
		"--format",
		"json",
		"solution",
		"list",
	).Return(nil, errors.New("list solution error"))

	saptuneClient := saptune.NewSaptuneClient(suite.mockExecutor, suite.logger)
	_, err := saptuneClient.ListSolution(ctx)

	suite.Error(err)
	suite.ErrorContains(err, "error executing saptune command: list solution error")
}

func (suite *SaptuneClientTestSuite) TestVerifySolution() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"saptune",
		"--format",
		"json",
		"solution",
		"verify",
	).Return([]byte("verify solution output"), nil)

	saptuneClient := saptune.NewSaptuneClient(suite.mockExecutor, suite.logger)
	verifySolutionOutput, err := saptuneClient.VerifySolution(ctx)

	suite.NoError(err)
	suite.Equal([]byte("verify solution output"), verifySolutionOutput)
}

func (suite *SaptuneClientTestSuite) TestVerifySolutionError() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"saptune",
		"--format",
		"json",
		"solution",
		"verify",
	).Return(nil, errors.New("verify solution error"))

	saptuneClient := saptune.NewSaptuneClient(suite.mockExecutor, suite.logger)
	_, err := saptuneClient.VerifySolution(ctx)

	suite.Error(err)
	suite.ErrorContains(err, "error executing saptune command: verify solution error")
}

func (suite *SaptuneClientTestSuite) TestListNote() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"saptune",
		"--format",
		"json",
		"note",
		"list",
	).Return([]byte("list note output"), nil)

	saptuneClient := saptune.NewSaptuneClient(suite.mockExecutor, suite.logger)
	listNoteOutput, err := saptuneClient.ListNote(ctx)

	suite.NoError(err)
	suite.Equal([]byte("list note output"), listNoteOutput)
}

func (suite *SaptuneClientTestSuite) TestListNoteError() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"saptune",
		"--format",
		"json",
		"note",
		"list",
	).Return(nil, errors.New("list note error"))

	saptuneClient := saptune.NewSaptuneClient(suite.mockExecutor, suite.logger)
	_, err := saptuneClient.ListNote(ctx)

	suite.Error(err)
	suite.ErrorContains(err, "error executing saptune command: list note error")
}

func (suite *SaptuneClientTestSuite) TestVerifyNote() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"saptune",
		"--format",
		"json",
		"note",
		"verify",
	).Return([]byte("verify note output"), nil)

	saptuneClient := saptune.NewSaptuneClient(suite.mockExecutor, suite.logger)
	verifyNoteOutput, err := saptuneClient.VerifyNote(ctx)

	suite.NoError(err)
	suite.Equal([]byte("verify note output"), verifyNoteOutput)
}

func (suite *SaptuneClientTestSuite) TestVerifyNoteError() {
	ctx := context.Background()

	suite.mockExecutor.On(
		"CombinedOutputContext",
		ctx,
		"saptune",
		"--format",
		"json",
		"note",
		"verify",
	).Return(nil, errors.New("verify note error"))

	saptuneClient := saptune.NewSaptuneClient(suite.mockExecutor, suite.logger)
	_, err := saptuneClient.VerifyNote(ctx)

	suite.Error(err)
	suite.ErrorContains(err, "error executing saptune command: verify note error")
}
