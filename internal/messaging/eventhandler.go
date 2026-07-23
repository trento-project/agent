// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"

	"github.com/trento-project/agent/v3/internal/factsengine/gatherers"
	"github.com/trento-project/agent/v3/internal/operations/operator"
)

type AnyRegistry interface {
	operator.Registry | gatherers.Registry
}

type EventHandler func(name string, event []byte) error

func MakeEventHandler[R AnyRegistry](
	ctx context.Context,
	agentID string,
	adapter Adapter,
	registry R,
	handleEvent func(ctx context.Context, request []byte, agentID string, adapter Adapter, registry R) error,
) EventHandler {
	return func(_ string, event []byte) error {
		return handleEvent(ctx, event, agentID, adapter, registry)
	}
}
