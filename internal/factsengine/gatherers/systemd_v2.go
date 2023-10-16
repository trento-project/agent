package gatherers

import (
	"context"
	"encoding/json"

	"github.com/coreos/go-systemd/v22/dbus"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const (
	SystemDGathererNameV2 = "systemd@v2"
)

// nolint:gochecknoglobals
var (
	SystemDUnitError = entities.FactGatheringError{
		Type:    "systemd-unit-error",
		Message: "error getting systemd unit properties",
	}

	SystemDDecodingError = entities.FactGatheringError{
		Type:    "systemd-decoding-error",
		Message: "error decoding systemd unit status",
	}
)

type SystemDGathererV2 struct {
	dbusConnnector DbusConnector
	initialized    bool
}

type systemdUnitStatus struct {
	ActiveState      interface{} `json:"active_state"`
	Description      interface{} `json:"description"`
	ID               interface{} `json:"id"`
	LoadState        interface{} `json:"load_state"`
	NeedDaemonReload interface{} `json:"need_daemon_reload"`
	UnitFilePreset   interface{} `json:"unit_file_preset"`
	UnitFileState    interface{} `json:"unit_file_state"`
}

func NewDefaultSystemDGathererV2() *SystemDGathererV2 {
	ctx := context.Background()
	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		log.Errorf("Error initializing dbus: %s", err)
		return &SystemDGathererV2{
			dbusConnnector: nil,
			initialized:    false,
		}
	}

	return NewSystemDGathererV2(conn, true)
}

func NewSystemDGathererV2(conn DbusConnector, initialized bool) *SystemDGathererV2 {
	return &SystemDGathererV2{
		dbusConnnector: conn,
		initialized:    initialized,
	}
}

func (g *SystemDGathererV2) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting %s v2 facts gathering process", SystemDGathererName)

	if !g.initialized {
		return facts, &SystemDNotInitializedError
	}

	ctx := context.Background()

	for _, factReq := range factsRequests {
		if len(factReq.Argument) == 0 {
			log.Error(SystemDMissingArgument.Message)
			fact := entities.NewFactGatheredWithError(factReq, &SystemDMissingArgument)
			facts = append(facts, fact)
		} else {
			properties, err := g.dbusConnnector.GetUnitPropertiesContext(ctx, factReq.Argument)
			if err != nil {
				gatheringError := SystemDUnitError.Wrap(err.Error())
				log.Error(gatheringError)
				facts = append(facts, entities.NewFactGatheredWithError(factReq, gatheringError))
				continue
			}

			var fact entities.Fact

			factValue, err := unitPropertiesToFactValue(properties)
			if err == nil {
				fact = entities.NewFactGatheredWithRequest(factReq, factValue)
			} else {
				gatheringError := SystemDDecodingError.Wrap(err.Error())
				log.Error(gatheringError)
				fact = entities.NewFactGatheredWithError(factReq, gatheringError)
			}
			facts = append(facts, fact)
		}
	}

	log.Infof("Requested %s v2 facts gathered", SystemDGathererName)
	return facts, nil
}

func unitPropertiesToFactValue(properties map[string]interface{}) (entities.FactValue, error) {
	unitStatus := systemdUnitStatus{
		ActiveState:      properties["ActiveState"],
		Description:      properties["Description"],
		ID:               properties["Id"],
		LoadState:        properties["LoadState"],
		NeedDaemonReload: properties["NeedDaemonReload"],
		UnitFilePreset:   properties["UnitFilePreset"],
		UnitFileState:    properties["UnitFileState"],
	}

	marshalled, err := json.Marshal(&unitStatus)
	if err != nil {
		return nil, err
	}

	var unmarshalled map[string]interface{}
	err = json.Unmarshal(marshalled, &unmarshalled)
	if err != nil {
		return nil, err
	}

	return entities.NewFactValue(unmarshalled, entities.WithStringConversion())
}
