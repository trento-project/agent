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

func (s *packageVersionConfigGatherer) Gather(packages []string) ([]*Fact, error) {
	var facts []*Fact
	log.Infof("Starting Package versions facts gathering process")

	for _, packageName := range packages {
		version, _ := exec.Command("rpm", "-q", "--qf", "%{VERSION}", packageName).Output()
		fact := &Fact{
			Name:  PackageVersionFactKey,
			Key:   packageName,
			Value: string(version),
		}

		facts = append(facts, fact)
	}

	log.Infof("Requested Package versions facts gathered")
	return facts, nil
}
