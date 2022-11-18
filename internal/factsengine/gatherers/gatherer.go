package gatherers

import (
	"github.com/trento-project/agent/internal/factsengine/entities"
)

type FactGatherer interface {
	Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error)
}

func StandardGatherers() map[string]FactGatherer {
	return map[string]FactGatherer{
		CorosyncConfGathererName:    NewDefaultCorosyncConfGatherer(),
		HostsFileGathererName:       NewDefaultHostsFileGatherer(),
		SystemDGathererName:         NewDefaultSystemDGatherer(),
		PackageVersionGathererName:  NewDefaultPackageVersionGatherer(),
		SBDConfigGathererName:       NewDefaultSBDGatherer(),
		CorosyncCmapCtlGathererName: NewDefaultCorosyncCmapctlGatherer(),
	}
}
