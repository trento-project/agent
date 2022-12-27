package gatherers

import (
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
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

	PackageVersionMissingArgument = entities.FactGatheringError{
		Type:    "package-version-missing-argument",
		Message: "missing required argument",
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
	log.Infof("Starting %s facts gathering process", PackageVersionGathererName)

	for _, factReq := range factsRequests {
		var fact entities.Fact
		if len(factReq.Argument) < 1 {
			log.Error(PackageVersionMissingArgument.Message)
			fact = entities.NewFactGatheredWithError(factReq, &PackageVersionMissingArgument)
			facts = append(facts, fact)
			continue
		}

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

	log.Infof("Requested %s facts gathered", PackageVersionGathererName)
	return facts, nil
}
