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
)

//go:generate mockery --name=DbusConnector
type DbusConnector interface {
	ListUnitsByNamesContext(ctx context.Context, units []string) ([]dbus.UnitStatus, error)
}

type SystemDGatherer struct {
	dbusConnnector DbusConnector
	initialized    bool
}

func NewDefaultSystemDGatherer() *SystemDGatherer {
	ctx := context.Background()
	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		log.Errorf("Error initializing dbus: %s", err)
		return &SystemDGatherer{
			dbusConnnector: nil,
			initialized:    false,
		}
	}

	return NewSystemDGatherer(conn, true)
}

func NewSystemDGatherer(conn DbusConnector, initialized bool) *SystemDGatherer {
	return &SystemDGatherer{
		dbusConnnector: conn,
		initialized:    initialized,
	}
}

func (g *SystemDGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting systemd state facts gathering process")

	if !g.initialized {
		return facts, &SystemDNotInitializedError
	}

	services := []string{}
	for _, factReq := range factsRequests {
		services = append(services, completeServiceName(factReq.Argument))
	}

	ctx := context.Background()

	states, err := g.dbusConnnector.ListUnitsByNamesContext(ctx, services)
	if err != nil {
		return facts, SystemDListUnitsError.Wrap(err.Error())
	}

	for index, factReq := range factsRequests {
		if states[index].Name == completeServiceName(factReq.Argument) {
			state := &entities.FactValueString{Value: states[index].ActiveState}
			fact := entities.NewFactGatheredWithRequest(factReq, state)
			facts = append(facts, fact)
		}
	}

	log.Infof("Requested systemd state facts gathered")
	return facts, nil
}

func completeServiceName(service string) string {
	return fmt.Sprintf("%s.service", service)
}
