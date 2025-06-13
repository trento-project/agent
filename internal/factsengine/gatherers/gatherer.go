package gatherers

import (
	"context"

	"github.com/trento-project/agent/internal/factsengine/factscache"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

// nolint:gochecknoglobals
var ImplementationError = entities.FactGatheringError{
	Type:    "implemetation-error",
	Message: "implementation error",
}

type FactGatherer interface {
	Gather(context context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error)
}

type FactGathererWithCache interface {
	SetCache(cache *factscache.FactsCache)
}

func StandardGatherers() FactGatherersTree {
	return FactGatherersTree{
		AscsErsClusterGathererName: map[string]FactGatherer{
			"v1": NewDefaultAscsErsClusterGatherer(),
		},
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
		FSUsageGathererName: map[string]FactGatherer{
			"v1": NewDefaultFSUsageGatherer(),
		},
		GroupsGathererName: map[string]FactGatherer{
			"v1": NewDefaultGroupsGatherer(),
		},
		HostsFileGathererName: map[string]FactGatherer{
			"v1": NewDefaultHostsFileGatherer(),
		},
		IniFilesGathererName: map[string]FactGatherer{
			"v1": NewDefaultIniFilesGatherer(),
		},
		MountInfoGathererName: map[string]FactGatherer{
			"v1": NewDefaultMountInfoGatherer(),
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
		ProductsGathererName: map[string]FactGatherer{
			"v1": NewDefaultProductsGatherer(),
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
		SudoersGathererName: map[string]FactGatherer{
			"v1": NewDefaultSudoersGatherer(),
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
