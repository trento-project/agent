package gatherers

import (
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

type FactGatherer interface {
	Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error)
}

func StandardGatherers() map[string]FactGatherer {
	return map[string]FactGatherer{
		CibAdminGathererName:        NewDefaultCibAdminGatherer(),
		CorosyncCmapCtlGathererName: NewDefaultCorosyncCmapctlGatherer(),
		CorosyncConfGathererName:    NewDefaultCorosyncConfGatherer(),
		HostsFileGathererName:       NewDefaultHostsFileGatherer(),
		SystemDGathererName:         NewDefaultSystemDGatherer(),
		PackageVersionGathererName:  NewDefaultPackageVersionGatherer(),
		PasswdGathererName:          NewDefaultPasswdGatherer(),
		SBDConfigGathererName:       NewDefaultSBDGatherer(),
		SBDDumpGathererName:         NewDefaultSBDDumpGatherer(),
		SapHostCtrlGathererName:     NewDefaultSapHostCtrlGatherer(),
		VerifyPasswordGathererName:  NewDefaultPasswordGatherer(),
	}
}
