package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/trento-project/agent/internal/core/sapsystem"
)

const SAPDiscoveryID string = "sap_system_discovery"
const SAPDiscoveryMinPeriod time.Duration = 1 * time.Second

type SAPSystemsDiscovery struct {
}

func NewSAPSystemsDiscovery() Discoverer[sapsystem.SAPSystemsList] {
	return SAPSystemsDiscovery{}
}

func (d SAPSystemsDiscovery) Discover(ctx context.Context) (sapsystem.SAPSystemsList, string, error) {
	systems, err := sapsystem.NewDefaultSAPSystemsList(ctx)

	if err != nil {
		return nil, "", err
	}

	sysNames := systems.GetSIDsString()
	if sysNames != "" {

		return systems, fmt.Sprintf("SAP system(s) with ID: %s discovered", sysNames), nil
	}

	return systems, "No SAP system discovered on this host", nil
}
