package gatherers

import (
	"os"

	"github.com/hashicorp/go-envparse"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const (
	OSReleaseGathererName = "os-release"
	OSReleaseFilePath     = "/etc/os-release"
)

// nolint:gochecknoglobals
var (
	OSReleaseFileError = entities.FactGatheringError{
		Type:    "os-release-file-error",
		Message: "error reading /etc/os-release file",
	}

	OSReleaseDecodingError = entities.FactGatheringError{
		Type:    "os-release-decoding-error",
		Message: "error decoding file content",
	}
)

type OSReleaseGatherer struct {
	osReleaseFilePath string
}

func NewDefaultOSReleaseGatherer() *OSReleaseGatherer {
	return NewOSReleaseGatherer(OSReleaseFilePath)
}

func NewOSReleaseGatherer(path string) *OSReleaseGatherer {
	return &OSReleaseGatherer{
		osReleaseFilePath: path,
	}
}

func (g *OSReleaseGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting %s facts gathering process", OSReleaseGathererName)

	file, err := os.Open(g.osReleaseFilePath)
	if err != nil {
		log.Error(err)
		return facts, OSReleaseFileError.Wrap(err.Error())
	}
	defer file.Close()

	osRelease, err := envparse.Parse(file)
	if err != nil {
		log.Error(err)
		return facts, OSReleaseDecodingError.Wrap(err.Error())
	}

	osReleaseFactValue := mapToFactValue(osRelease)
	if err != nil {
		log.Error(err)
		return facts, OSReleaseDecodingError.Wrap(err.Error())
	}

	for _, requestedFact := range factsRequests {
		fact := entities.NewFactGatheredWithRequest(requestedFact, osReleaseFactValue)
		facts = append(facts, fact)
	}

	log.Infof("Requested %s facts gathered", OSReleaseGathererName)
	return facts, nil
}

func mapToFactValue(inputMap map[string]string) entities.FactValue {
	factValueMap := make(map[string]entities.FactValue)

	for key, value := range inputMap {
		factValueMap[key] = &entities.FactValueString{Value: value}
	}

	return &entities.FactValueMap{Value: factValueMap}
}
