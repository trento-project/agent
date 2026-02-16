package operator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/trento-project/agent/internal/core/saptune"
	"github.com/trento-project/agent/pkg/utils"
)

const SaptuneChangeSolutionOperatorName = "saptunechangesolution"

type SaptuneChangeSolutionOption Option[SaptuneChangeSolution]

// SaptuneChangeSolution is an operator responsible for changing a saptune solution.
//
// It requires the same kind of argument needed for SaptuneApplySolution: a map containing a key named "solution".
//
// # Execution Phases
//
// - PLAN:
//   Same as SaptuneApplySolutionOption.
//
// - COMMIT:
//   If there is no other solution already applied, an error is raised,
//   effectively allowing a transition from a solution to another but not from no solution to some solution.
// 	 If otherwise there is a solution applied that is not the currently applied one the saptune command to change the
//   solution will be executed.
//
// - VERIFY:
//   The operator verifies whether the solution has been correctly changed to the requested one.
//   If not, an error is raised.
//
//   On success, the current state of the applied solution is collected as the "after" diff.
//
// - ROLLBACK:
//   If an error occurs during the COMMIT or VERIFY phase, the saptune solution is changed back to the initially
//   applied one.

type SaptuneChangeSolution struct {
	baseOperator
	saptune         saptune.Saptune
	parsedArguments *saptuneSolutionArguments
}

func WithSaptuneClientChange(saptuneClient saptune.Saptune) SaptuneChangeSolutionOption {
	return func(o *SaptuneChangeSolution) {
		o.saptune = saptuneClient
	}
}

func NewSaptuneChangeSolution(
	arguments Arguments,
	operationID string,
	options Options[SaptuneChangeSolution],
) *Executor {
	saptuneChange := &SaptuneChangeSolution{
		baseOperator: newBaseOperator(
			SaptuneChangeSolutionOperatorName,
			operationID,
			arguments,
			options.BaseOperatorOptions...,
		),
	}

	saptuneChange.saptune = saptune.NewSaptuneClient(
		utils.Executor{},
		saptuneChange.logger,
	)

	for _, opt := range options.OperatorOptions {
		opt(saptuneChange)
	}

	return &Executor{
		phaser:      saptuneChange,
		operationID: operationID,
		logger:      saptuneChange.logger,
	}
}

func (sc *SaptuneChangeSolution) plan(ctx context.Context) (bool, error) {
	opArguments, err := parseSaptuneSolutionArguments(sc.arguments)
	if err != nil {
		return false, err
	}
	sc.parsedArguments = opArguments

	if err = sc.saptune.CheckVersionSupport(ctx); err != nil {
		return false, err
	}

	initiallyAppliedSolution, err := sc.saptune.GetAppliedSolution(ctx)
	if err != nil {
		return false, err
	}

	sc.resources[beforeDiffField] = initiallyAppliedSolution

	if sc.parsedArguments.solution == initiallyAppliedSolution {
		sc.logger.Info("solution is already applied. Nothing to change, skipping operation",
			"solution", sc.parsedArguments.solution)
		sc.resources[afterDiffField] = initiallyAppliedSolution
		return true, nil
	}

	return false, nil
}

func (sc *SaptuneChangeSolution) commit(ctx context.Context) error {
	initiallyAppliedSolution, _ := sc.resources[beforeDiffField].(string)

	if initiallyAppliedSolution == "" {
		return fmt.Errorf(
			"cannot change solution to %s because no solution is currently applied",
			sc.parsedArguments.solution,
		)
	}

	return sc.saptune.ChangeSolution(ctx, sc.parsedArguments.solution)
}

func (sc *SaptuneChangeSolution) verify(ctx context.Context) error {
	initiallyAppliedSolution, _ := sc.resources[beforeDiffField].(string)

	if sc.parsedArguments.solution == initiallyAppliedSolution {
		sc.resources[afterDiffField] = initiallyAppliedSolution
		return nil
	}

	appliedSolution, err := sc.saptune.GetAppliedSolution(ctx)
	if err != nil {
		return err
	}

	if appliedSolution != sc.parsedArguments.solution {
		return fmt.Errorf(
			"verify saptune apply failing, the solution %s was not applied in commit phase",
			sc.parsedArguments.solution,
		)
	}
	sc.resources[afterDiffField] = appliedSolution
	return nil
}

func (sc *SaptuneChangeSolution) rollback(ctx context.Context) error {
	initiallyAppliedSolution, _ := sc.resources[beforeDiffField].(string)

	if initiallyAppliedSolution == "" {
		return nil
	}

	sc.logger.Info("Changing solution to the initially applied one",
		"appliedSolution", initiallyAppliedSolution)
	return sc.saptune.ChangeSolution(ctx, initiallyAppliedSolution)
}

//	operationDiff needs to be refactored, ignoring duplication issues for now
//
// nolint: dupl
func (sc *SaptuneChangeSolution) operationDiff(_ context.Context) map[string]any {
	diff := make(map[string]any)

	beforeSolution, ok := sc.resources[beforeDiffField].(string)
	if !ok {
		panic(fmt.Sprintf("invalid beforeSolution value: cannot parse '%s' to string",
			sc.resources[beforeDiffField]))
	}

	afterSolution, ok := sc.resources[afterDiffField].(string)
	if !ok {
		panic(fmt.Sprintf("invalid afterSolution value: cannot parse '%s' to string",
			sc.resources[afterDiffField]))
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
