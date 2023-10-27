package gatherers

import (
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

type FactGatherer interface {
	Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error)
}

func StandardGatherers() FactGatherersTree {
	return FactGatherersTree{
		CibAdminGathererName: map[string]FactGatherer{
			"v1": NewDefaultCibAdminGatherer(),
		},
		CorosyncCmapCtlGathererName: map[string]FactGatherer{
			"v1": NewDefaultCorosyncCmapctlGatherer(),
		},
		CorosyncConfGathererName: map[string]FactGatherer{
			"v1": NewDefaultCorosyncConfGatherer(),
		},
		DirScanGathererName: map[string]FactGatherer{
			"v1": NewDefaultDirScanGatherer(),
		},
		DispWorkGathererName: map[string]FactGatherer{
			"v1": NewDefaultDispWorkGatherer(),
		},
		FstabGathererName: map[string]FactGatherer{
			"v1": NewDefaultFstabGatherer(),
		},
		GroupsGathererName: map[string]FactGatherer{
			"v1": NewDefaultGroupsGatherer(),
		},
		HostsFileGathererName: map[string]FactGatherer{
			"v1": NewDefaultHostsFileGatherer(),
		},
		OSReleaseGathererName: map[string]FactGatherer{
			"v1": NewDefaultOSReleaseGatherer(),
		},
		PackageVersionGathererName: map[string]FactGatherer{
			"v1": NewDefaultPackageVersionGatherer(),
		},
		PasswdGathererName: map[string]FactGatherer{
			"v1": NewDefaultPasswdGatherer(),
		},
		SapControlGathererName: map[string]FactGatherer{
			"v1": NewDefaultSapControlGatherer(),
		},
		SapHostCtrlGathererName: map[string]FactGatherer{
			"v1": NewDefaultSapHostCtrlGatherer(),
		},
		SapInstanceHostnameResolverGathererName: map[string]FactGatherer{
			"v1": NewDefaultSapInstanceHostnameResolverGatherer(),
		},
		SapProfilesGathererName: map[string]FactGatherer{
			"v1": NewDefaultSapProfilesGatherer(),
		},
		SapServicesGathererName: map[string]FactGatherer{
			"v1": NewDefaultSapServicesGatherer(),
		},
		SaptuneGathererName: map[string]FactGatherer{
			"v1": NewDefaultSaptuneGatherer(),
		},
		SBDConfigGathererName: map[string]FactGatherer{
			"v1": NewDefaultSBDGatherer(),
		},
		SBDDumpGathererName: map[string]FactGatherer{
			"v1": NewDefaultSBDDumpGatherer(),
		},
		SysctlGathererName: map[string]FactGatherer{
			"v1": NewDefaultSysctlGatherer(),
		},
		SystemDGathererName: map[string]FactGatherer{
			"v1": NewDefaultSystemDGatherer(),
			"v2": NewDefaultSystemDGathererV2(),
		},
		VerifyPasswordGathererName: map[string]FactGatherer{
			"v1": NewDefaultPasswordGatherer(),
		},
	}
}
