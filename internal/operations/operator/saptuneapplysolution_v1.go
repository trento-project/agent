package operator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/trento-project/agent/internal/core/saptune"
	"github.com/trento-project/agent/pkg/utils"
)

type saptuneSolutionArguments struct {
	solution string
}

func parseSaptuneSolutionArguments(rawArguments Arguments) (*saptuneSolutionArguments, error) {
	argument, found := rawArguments["solution"]
	if !found {
		return nil, errors.New("argument solution not provided, could not use the operator")
	}

	solution, ok := argument.(string)
	if !ok {
		return nil, fmt.Errorf(
			"could not parse solution argument as string, argument provided: %v",
			argument,
		)
	}

	if solution == "" {
		return nil, errors.New("solution argument is empty")
	}

	return &saptuneSolutionArguments{
		solution: solution,
	}, nil
}

const SaptuneApplySolutionOperatorName = "saptuneapplysolution"

type SaptuneApplySolutionOption Option[SaptuneApplySolution]

// SaptuneApplySolution is an operator responsible for applying a saptune solution.
//
// The operator requires an argument in the form of a map containing a key named "solution".
// This value will be passed to the saptune command-line tool.
//
// All considerations related to applying a solution using the saptune CLI apply here as well.
//
// # Execution Phases
//
// - PLAN:
//   The operator checks for the presence of the saptune binary and verifies its version.
//   The minimum required version is 3.1.0. If saptune is not installed or the version does not meet the minimum
//   requirement, the operation will fail.
//
//   The initially applied solution, if any, is collected as the "before" diff.
//   The operator checks if the requested solution is already applied. If it is, no action is taken,
//   ensuring idempotency without returning an error.
//
// - COMMIT:
//   If there is any other solution already applied, an error is raised, because only one solution can be
//   applied at a time.
// 	 If otherwise there is no solution applied the saptune command to apply the solution will be executed.
//
// - VERIFY:
//   The operator verifies whether the solution has been correctly applied to the system.
//   If not, an error is raised. If successful, the current state of the applied solution is collected as
//   the "after" diff.
//
// - ROLLBACK:
//   If an error occurs during the COMMIT or VERIFY phase, the saptune revert command is executed
//   to undo the applied solution.

type SaptuneApplySolution struct {
	baseOperator
	saptune         saptune.Saptune
	parsedArguments *saptuneSolutionArguments
}

type saptuneOperationDiffOutput struct {
	Solution string `json:"solution"`
}

func WithSaptuneClientApply(saptuneClient saptune.Saptune) SaptuneApplySolutionOption {
	return func(o *SaptuneApplySolution) {
		o.saptune = saptuneClient
	}
}

func NewSaptuneApplySolution(
	arguments Arguments,
	operationID string,
	options Options[SaptuneApplySolution],
) *Executor {
	saptuneApply := &SaptuneApplySolution{
		baseOperator: newBaseOperator(
			SaptuneApplySolutionOperatorName, operationID, arguments, options.BaseOperatorOptions...,
		),
	}

	saptuneApply.saptune = saptune.NewSaptuneClient(
		utils.Executor{},
		saptuneApply.logger,
	)

	for _, opt := range options.OperatorOptions {
		opt(saptuneApply)
	}

	return &Executor{
		phaser:      saptuneApply,
		operationID: operationID,
		logger:      saptuneApply.logger,
	}
}

func (sa *SaptuneApplySolution) plan(ctx context.Context) (bool, error) {
	opArguments, err := parseSaptuneSolutionArguments(sa.arguments)
	if err != nil {
		return false, err
	}
	sa.parsedArguments = opArguments

	if err = sa.saptune.CheckVersionSupport(ctx); err != nil {
		return false, err
	}

	initiallyAppliedSolution, err := sa.saptune.GetAppliedSolution(ctx)
	if err != nil {
		return false, err
	}

	sa.resources[beforeDiffField] = initiallyAppliedSolution

	if sa.parsedArguments.solution == initiallyAppliedSolution {
		sa.logger.Info("solution is already applied, skipping operation", "solution", sa.parsedArguments.solution)
		sa.resources[afterDiffField] = initiallyAppliedSolution
		return true, nil
	}

	return false, nil
}

func (sa *SaptuneApplySolution) commit(ctx context.Context) error {
	initiallyAppliedSolution, _ := sa.resources[beforeDiffField].(string)

	if initiallyAppliedSolution != "" {
		return fmt.Errorf(
			"cannot apply solution %s because another solution %s is already applied",
			sa.parsedArguments.solution,
			initiallyAppliedSolution,
		)
	}

	return sa.saptune.ApplySolution(ctx, sa.parsedArguments.solution)
}

func (sa *SaptuneApplySolution) verify(ctx context.Context) error {
	appliedSolution, err := sa.saptune.GetAppliedSolution(ctx)
	if err != nil {
		return err
	}

	if appliedSolution != sa.parsedArguments.solution {
		return fmt.Errorf(
			"verify saptune apply failing, the solution %s was not applied in commit phase",
			sa.parsedArguments.solution,
		)
	}
	sa.resources[afterDiffField] = appliedSolution
	return nil
}

func (sa *SaptuneApplySolution) rollback(ctx context.Context) error {
	initiallyAppliedSolution, _ := sa.resources[beforeDiffField].(string)

	if initiallyAppliedSolution != "" {
		return nil
	}

	return sa.saptune.RevertSolution(ctx, sa.parsedArguments.solution)
}

//	operationDiff needs to be refactored, ignoring duplication issues for now
//
// nolint: dupl
func (sa *SaptuneApplySolution) operationDiff(_ context.Context) map[string]any {
	diff := make(map[string]any)

	beforeSolution, ok := sa.resources[beforeDiffField].(string)
	if !ok {
		panic(fmt.Sprintf("invalid beforeSolution value: cannot parse '%s' to string",
			sa.resources[beforeDiffField]))
	}

	afterSolution, ok := sa.resources[afterDiffField].(string)
	if !ok {
		panic(fmt.Sprintf("invalid afterSolution value: cannot parse '%s' to string",
			sa.resources[afterDiffField]))
	}

	beforeDiffOutput := saptuneOperationDiffOutput{
		Solution: beforeSolution,
	}
	before, err := json.Marshal(beforeDiffOutput)
	if err != nil {
		panic(fmt.Sprintf("error marshalling before diff output: %v", err))
	}
	diff[beforeDiffField] = string(before)

	afterDiffOutput := saptuneOperationDiffOutput{
		Solution: afterSolution,
	}
	after, err := json.Marshal(afterDiffOutput)
	if err != nil {
		panic(fmt.Sprintf("error marshalling after diff output: %v", err))
	}
	diff[afterDiffField] = string(after)

	return diff
}
