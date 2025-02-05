package gatherers

import (
	"context"
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
	log "github.com/sirupsen/logrus"
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

//go:generate mockery --name=DbusConnector
type DbusConnector interface {
	GetUnitPropertiesContext(ctx context.Context, unit string) (map[string]interface{}, error)
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

func (g *SystemDGatherer) Gather(ctx context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting %s facts gathering process", SystemDGathererName)

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
	if err != nil {
		return facts, SystemDListUnitsError.Wrap(err.Error())
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	for index, factReq := range factsRequests {
		if len(factReq.Argument) == 0 {
			log.Error(SystemDMissingArgument.Message)
			fact := entities.NewFactGatheredWithError(factReq, &SystemDMissingArgument)
			facts = append(facts, fact)
		} else if states[index].Name == completeServiceName(factReq.Argument) {
			state := &entities.FactValueString{Value: states[index].ActiveState}
			fact := entities.NewFactGatheredWithRequest(factReq, state)
			facts = append(facts, fact)
		}
	}

	log.Infof("Requested %s facts gathered", SystemDGathererName)
	return facts, nil
}

func completeServiceName(service string) string {
	return fmt.Sprintf("%s.service", service)
}
