package operator

import (
	"context"
	"log/slog"

	"github.com/trento-project/agent/pkg/utils"
)

const (
	beforeDiffField = "before"
	afterDiffField  = "after"
)

type BaseOperatorOption Option[baseOperator]

func WithCustomLogger(logger *slog.Logger) BaseOperatorOption {
	return func(b *baseOperator) {
		b.logger = logger
	}
}

type baseOperator struct {
	arguments Arguments
	resources map[string]any
	logger    *slog.Logger
}

func newBaseOperator(
	name string,
	operationID string,
	arguments Arguments,
	options ...BaseOperatorOption,
) baseOperator {
	base := &baseOperator{
		arguments: arguments,
		resources: make(map[string]any),
		logger:    utils.NewDefaultLogger("info"),
	}

	for _, opt := range options {
		opt(base)
	}

	base.logger = base.logger.
		With("operation_id", operationID).
		With("operator_name", name)

	return *base
}

func (b *baseOperator) after(_ context.Context) {}
