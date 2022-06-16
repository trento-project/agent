package facts

import (
	"os/exec"

	log "github.com/sirupsen/logrus"
)

const (
	PackageVersionFactKey = "package_version"
)

type packageVersionConfigGatherer struct {
}

func NewPackageVersionConfigGatherer() *packageVersionConfigGatherer {
	return &packageVersionConfigGatherer{}
}

func (s *packageVersionConfigGatherer) Gather(factsRequests []FactRequest) ([]*Fact, error) {
	var facts []*Fact
	log.Infof("Starting Package versions facts gathering process")

	for _, factReq := range factsRequests {
		version, _ := exec.Command("rpm", "-q", "--qf", "%{VERSION}", factReq.Name).Output()
		fact := &Fact{
			Name:  PackageVersionFactKey,
			Key:   factReq.Name,
			Value: string(version),
			Alias: factReq.Alias,
		}

		facts = append(facts, fact)
	}

	log.Infof("Requested Package versions facts gathered")
	return facts, nil
}
