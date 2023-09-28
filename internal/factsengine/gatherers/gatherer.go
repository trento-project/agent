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
		GroupsGathererName:          NewDefaultGroupsGatherer(),
		HostsFileGathererName:       NewDefaultHostsFileGatherer(),
		PackageVersionGathererName:  NewDefaultPackageVersionGatherer(),
		PasswdGathererName:          NewDefaultPasswdGatherer(),
		SapHostCtrlGathererName:     NewDefaultSapHostCtrlGatherer(),
		SaptuneGathererName:         NewDefaultSaptuneGatherer(),
		SBDConfigGathererName:       NewDefaultSBDGatherer(),
		SBDDumpGathererName:         NewDefaultSBDDumpGatherer(),
		SystemDGathererName:         NewDefaultSystemDGatherer(),
		VerifyPasswordGathererName:  NewDefaultPasswordGatherer(),
	}
}
