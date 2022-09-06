package gatherers

import (
	"context"
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/entities"
)

const (
	SystemDGathererName = "systemd"
)

var (
	SystemDNotInitializedError = entities.FactGatheringError{ // nolint
		Type:    "systemd-gatherer-not-initialized",
		Message: "systemd gatherer not initialized properly",
	}

	SystemDListUnitsError = entities.FactGatheringError{ // nolint
		Type:    "systemd-list-units-error",
		Message: "error getting unit states",
	}
)

//go:generate mockery --name=DbusConnector
type DbusConnector interface {
	ListUnitsByNamesContext(ctx context.Context, units []string) ([]dbus.UnitStatus, error)
}

type SystemDGatherer struct {
	dbusConnnector DbusConnector
	initialized    bool
}

func NewSystemDGatherer() *SystemDGatherer {
	ctx := context.Background()
	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		log.Errorf("Error initializing dbus: %s", err)
		return &SystemDGatherer{
			dbusConnnector: nil,
			initialized:    false,
		}
	}

	return &SystemDGatherer{
		dbusConnnector: conn,
		initialized:    true,
	}
}

func (g *SystemDGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.FactsGatheredItem, error) {
	facts := []entities.FactsGatheredItem{}
	log.Infof("Starting systemd state facts gathering process")

	if !g.initialized {
		gatheringError := SystemDNotInitializedError
		log.Errorf(gatheringError.Error())
		return entities.NewFactsGatheredListWithError(factsRequests, &gatheringError), nil
	}

	services := []string{}
	for _, factReq := range factsRequests {
		services = append(services, completeServiceName(factReq.Argument))
	}

	ctx := context.Background()

	states, err := g.dbusConnnector.ListUnitsByNamesContext(ctx, services)
	if err != nil {
		gatheringError := SystemDListUnitsError.Wrap(err.Error())
		log.Errorf(gatheringError.Error())
		return entities.NewFactsGatheredListWithError(factsRequests, &gatheringError), nil
	}

	for index, factReq := range factsRequests {
		if states[index].Name == completeServiceName(factReq.Argument) {
			fact := entities.NewFactGatheredWithRequest(factReq, states[index].ActiveState)
			facts = append(facts, fact)
		}
	}

	log.Infof("Requested systemd state facts gathered")
	return facts, nil
}

func completeServiceName(service string) string {
	return fmt.Sprintf("%s.service", service)
}
