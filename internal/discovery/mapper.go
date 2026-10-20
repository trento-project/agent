// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: GPL-3.0-or-later

package discovery

import (
	"github.com/trento-project/contracts/go/pkg/events"
)

//nolint:revive
type DiscoveryRequested struct {
	DiscoveryType string
	Targets       []string
}

//nolint:revive
func DiscoveryRequestedFromEvent(event []byte) (*DiscoveryRequested, error) {
	var discoveryRequested events.DiscoveryRequested

	err := events.FromEvent(event, &discoveryRequested, events.WithExpirationCheck())
	if err != nil {
		return nil, err
	}

	return &DiscoveryRequested{
		DiscoveryType: discoveryRequested.GetDiscoveryType(),
		Targets:       discoveryRequested.GetTargets(),
	}, nil
}
