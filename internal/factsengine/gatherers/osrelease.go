package gatherers

import (
	"context"
	"os"

	"log/slog"

	"github.com/hashicorp/go-envparse"
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

func (g *OSReleaseGatherer) Gather(ctx context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	slog.Info("Starting facts gathering process", "gatherer", OSReleaseGathererName)

	file, err := os.Open(g.osReleaseFilePath)
	if err != nil {
		slog.Error("Error opening os-release file", "error", err.Error())
		return facts, OSReleaseFileError.Wrap(err.Error())
	}
	defer func() {
		err := file.Close()
		if err != nil {
			slog.Error("could not close os-release file", "path", g.osReleaseFilePath, "error", err.Error())
		}
	}()

	osRelease, err := envparse.Parse(file)
	if err != nil {
		slog.Error("Error decoding os-release file content", "error", err.Error())
		return facts, OSReleaseDecodingError.Wrap(err.Error())
	}

	osReleaseFactValue := mapOSReleaseToFactValue(osRelease)
	if err != nil {
		slog.Error("Error decoding os-release file content", "error", err.Error())
		return facts, OSReleaseDecodingError.Wrap(err.Error())
	}

	for _, requestedFact := range factsRequests {
		fact := entities.NewFactGatheredWithRequest(requestedFact, osReleaseFactValue)
		facts = append(facts, fact)
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	slog.Info("Requested facts gathered", "gatherer", OSReleaseGathererName)
	return facts, nil
}

func mapOSReleaseToFactValue(inputMap map[string]string) entities.FactValue {
	factValueMap := make(map[string]entities.FactValue)

	for key, value := range inputMap {
		factValueMap[key] = &entities.FactValueString{Value: value}
	}

	return &entities.FactValueMap{Value: factValueMap}
}
