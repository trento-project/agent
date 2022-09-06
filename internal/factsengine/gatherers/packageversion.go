package gatherers

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/entities"
)

const (
	PackageVersionGathererName = "package_version"
)

var PackageNotFoundError = entities.FactGatheringError{ // nolint
	Type:    "package_not_found",
	Message: "package not found",
}

type PackageVersionGatherer struct {
	executor CommandExecutor
}

func NewDefaultPackageVersionGatherer() *PackageVersionGatherer {
	return NewPackageVersionGatherer(Executor{})
}

func NewPackageVersionGatherer(executor CommandExecutor) *PackageVersionGatherer {
	return &PackageVersionGatherer{
		executor: executor,
	}
}

func (g *PackageVersionGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.FactsGatheredItem, error) {
	facts := []entities.FactsGatheredItem{}
	log.Infof("Starting Package versions facts gathering process")

	for _, factReq := range factsRequests {
		var fact entities.FactsGatheredItem

		version, err := g.executor.Exec(
			"rpm", "-q", "--qf", "%{VERSION}", factReq.Argument)
		if err == nil {
			fact = entities.NewFactGatheredWithRequest(factReq, string(version))
		} else {
			gatheringError := PackageNotFoundError.Wrap(strings.TrimSuffix(string(version), "\n"))
			log.Errorf(gatheringError.Error())
			fact = entities.NewFactGatheredWithError(factReq, &gatheringError)
		}

		facts = append(facts, fact)
	}

	log.Infof("Requested Package versions facts gathered")
	return facts, nil
}
