// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: GPL-3.0-or-later

package operator

import (
	"context"
)

type PhaseName string

type (
	Arguments     map[string]any
	Option[T any] func(*T)
)

const (
	PLAN     PhaseName = "PLAN"
	COMMIT   PhaseName = "COMMIT"
	VERIFY   PhaseName = "VERIFY"
	ROLLBACK PhaseName = "ROLLBACK"
)

type Operator interface {
	Run(ctx context.Context) *ExecutionReport
}

type Options[T any] struct {
	BaseOperatorOptions []BaseOperatorOption
	OperatorOptions     []Option[T]
}
