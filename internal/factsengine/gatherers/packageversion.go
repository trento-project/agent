package gatherers

import (
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/entities"
	"github.com/trento-project/agent/internal/utils"
)

const (
	PackageVersionGathererName = "package_version"
)

// nolint:gochecknoglobals
var (
	PackageVersionCommandError = entities.FactGatheringError{
		Type:    "package-version-cmd-error",
		Message: "error getting version of package",
	}
)

type PackageVersionGatherer struct {
	executor utils.CommandExecutor
}

func NewDefaultPackageVersionGatherer() *PackageVersionGatherer {
	return NewPackageVersionGatherer(utils.Executor{})
}

func NewPackageVersionGatherer(executor utils.CommandExecutor) *PackageVersionGatherer {
	return &PackageVersionGatherer{
		executor: executor,
	}
}

func (g *PackageVersionGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting Package versions facts gathering process")

	for _, factReq := range factsRequests {
		var fact entities.Fact
		version, err := g.executor.Exec(
			"rpm", "-q", "--qf", "%{VERSION}", factReq.Argument)
		if err != nil {
			gatheringError := PackageVersionCommandError.Wrap(factReq.Argument)
			log.Error(gatheringError)
			fact = entities.NewFactGatheredWithError(factReq, gatheringError)
		} else {
			fact = entities.NewFactGatheredWithRequest(factReq, &entities.FactValueString{Value: (string(version))})
		}

		facts = append(facts, fact)
	}

	log.Infof("Requested Package versions facts gathered")
	return facts, nil
}
