package gatherers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/trento-project/agent/internal/core/hosts/systemd"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const (
	SystemDGathererName = "systemd"
)

// nolint:gochecknoglobals
var (
	SystemDNotInitializedError = entities.FactGatheringError{
		Type:    "systemd-dbus-not-initialized",
		Message: "systemd gatherer not initialized properly",
	}

	SystemDListUnitsError = entities.FactGatheringError{
		Type:    "systemd-list-units-error",
		Message: "error getting unit states",
	}

	SystemDMissingArgument = entities.FactGatheringError{
		Type:    "systemd-missing-argument",
		Message: "missing required argument",
	}
)

type SystemDGatherer struct {
	dbusConnnector systemd.DbusConnector
	initialized    bool
}

func NewDefaultSystemDGatherer() *SystemDGatherer {
	ctx := context.Background()
	conn, err := systemd.NewDbusConnector(ctx)
	if err != nil {
		slog.Error("Error initializing dbus", "error", err)
		return &SystemDGatherer{
			dbusConnnector: nil,
			initialized:    false,
		}
	}

	return NewSystemDGatherer(conn, true)
}

func NewSystemDGatherer(conn systemd.DbusConnector, initialized bool) *SystemDGatherer {
	return &SystemDGatherer{
		dbusConnnector: conn,
		initialized:    initialized,
	}
}

func (g *SystemDGatherer) Gather(ctx context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	slog.Info("Starting facts gathering process", "gatherer", SystemDGathererName)

	if !g.initialized {
		return facts, &SystemDNotInitializedError
	}

	services := []string{}
	for _, factReq := range factsRequests {
		if len(factReq.Argument) == 0 {
			continue
		}
		services = append(services, completeServiceName(factReq.Argument))
	}

	states, err := g.dbusConnnector.ListUnitsByNamesContext(ctx, services)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		return facts, SystemDListUnitsError.Wrap(err.Error())
	}

	for index, factReq := range factsRequests {
		if len(factReq.Argument) == 0 {
			slog.Error(SystemDMissingArgument.Message)
			fact := entities.NewFactGatheredWithError(factReq, &SystemDMissingArgument)
			facts = append(facts, fact)
		} else if states[index].Name == completeServiceName(factReq.Argument) {
			state := &entities.FactValueString{Value: states[index].ActiveState}
			fact := entities.NewFactGatheredWithRequest(factReq, state)
			facts = append(facts, fact)
		}
	}

	slog.Info("Requested facts gathered", "gatherer", SystemDGathererName)
	return facts, nil
}

func completeServiceName(service string) string {
	return fmt.Sprintf("%s.service", service)
}
