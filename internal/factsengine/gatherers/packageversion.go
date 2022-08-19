package gatherers

import (
	log "github.com/sirupsen/logrus"
)

const (
	PackageVersionGathererName = "package_version"
)

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

func (g *PackageVersionGatherer) Gather(factsRequests []FactRequest) ([]Fact, error) {
	facts := []Fact{}
	log.Infof("Starting Package versions facts gathering process")

	for _, factReq := range factsRequests {
		version, err := g.executor.Exec(
			"rpm", "-q", "--qf", "%{VERSION}", factReq.Argument)
		if err != nil {
			// TODO: Decide together with Wanda how to deal with errors. `err` field in the fact result?
			log.Errorf("Error getting version of package: %s", factReq.Argument)
		}
		fact := NewFactWithRequest(factReq, string(version))
		facts = append(facts, fact)
	}

	log.Infof("Requested Package versions facts gathered")
	return facts, nil
}
