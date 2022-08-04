package gatherers

import (
	"context"
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	SystemDGathererName = "systemd"
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

func (g *SystemDGatherer) Gather(factsRequests []FactRequest) ([]Fact, error) {
	facts := []Fact{}
	log.Infof("Starting systemd state facts gathering process")

	if !g.initialized {
		return facts, errors.New("systemd gatherer not initialized properly")
	}

	services := []string{}
	for _, factReq := range factsRequests {
		services = append(services, completeServiceName(factReq.Argument))
	}

	ctx := context.Background()

	states, err := g.dbusConnnector.ListUnitsByNamesContext(ctx, services)
	if err != nil {
		return facts, errors.Wrap(err, "Error getting unit states")
	}

	for index, factReq := range factsRequests {
		if states[index].Name == completeServiceName(factReq.Argument) {
			fact := NewFactWithRequest(factReq, states[index].ActiveState)
			facts = append(facts, fact)
		}
	}

	log.Infof("Requested systemd state facts gathered")
	return facts, nil
}

func completeServiceName(service string) string {
	return fmt.Sprintf("%s.service", service)
}
