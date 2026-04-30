package gatherers

import (
	"context"
	"strings"

	"log/slog"

	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/version"
)

const (
	StatusGathererName  = "status"
	statusMachineIDPath = "/etc/machine-id"
)

// nolint:gochecknoglobals
var (
	trentoAgentNamespace = uuid.Must(uuid.Parse("fb92284e-aa5e-47f6-a883-bf9469e7a0dc"))

	StatusMachineIDError = entities.FactGatheringError{
		Type:    "status-machine-id-error",
		Message: "error reading machine ID",
	}
)

type StatusGatherer struct {
	fs            afero.Fs
	machineIDPath string
}

func NewDefaultStatusGatherer() *StatusGatherer {
	return NewStatusGatherer(afero.NewOsFs(), statusMachineIDPath)
}

func NewStatusGatherer(fs afero.Fs, machineIDPath string) *StatusGatherer {
	return &StatusGatherer{
		fs:            fs,
		machineIDPath: machineIDPath,
	}
}

func (g *StatusGatherer) Gather(ctx context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	slog.Info("Starting facts gathering process", "gatherer", StatusGathererName)

	machineIDBytes, err := afero.ReadFile(g.fs, g.machineIDPath)
	if err != nil {
		slog.Error("Error reading machine ID", "error", err)
		return facts, StatusMachineIDError.Wrap(err.Error())
	}

	machineID := strings.TrimSpace(string(machineIDBytes))
	agentID := uuid.NewSHA1(trentoAgentNamespace, []byte(machineID)).String()

	statusValue := &entities.FactValueMap{
		Value: map[string]entities.FactValue{
			"agent_id": &entities.FactValueString{Value: agentID},
			"version":  &entities.FactValueString{Value: version.Version},
		},
	}

	for _, requestedFact := range factsRequests {
		fact := entities.NewFactGatheredWithRequest(requestedFact, statusValue)
		facts = append(facts, fact)
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	slog.Info("Requested facts gathered", "gatherer", StatusGathererName)
	return facts, nil
}
