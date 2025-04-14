// nolint:revive
package discovery

import (
	"github.com/trento-project/contracts/go/pkg/events"
)

type DiscoveryRequested struct {
	DiscoveryType string
	Targets       []string
}

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
