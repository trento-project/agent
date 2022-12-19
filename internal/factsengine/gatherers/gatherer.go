package gatherers

import (
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

type FactGatherer interface {
	Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error)
}

func StandardGatherers() map[string]FactGatherer {
	return map[string]FactGatherer{
		CorosyncCmapCtlGathererName: NewDefaultCorosyncCmapctlGatherer(),
		CorosyncConfGathererName:    NewDefaultCorosyncConfGatherer(),
		HostsFileGathererName:       NewDefaultHostsFileGatherer(),
		SystemDGathererName:         NewDefaultSystemDGatherer(),
		PackageVersionGathererName:  NewDefaultPackageVersionGatherer(),
		SBDConfigGathererName:       NewDefaultSBDGatherer(),
		SBDDumpGathererName:         NewDefaultSBDDumpGatherer(),
	}
}
