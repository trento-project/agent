package gatherers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
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
		slog.Error("Error initializing dbus", "error", err)
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

func (g *SystemDGathererV2) Gather(ctx context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	slog.Info("Starting facts gathering process", "gatherer", SystemDGathererName, "version", "v2")

	if !g.initialized {
		return facts, &SystemDNotInitializedError
	}

	for _, factReq := range factsRequests {
		if len(factReq.Argument) == 0 {
			slog.Error(SystemDMissingArgument.Error())
			fact := entities.NewFactGatheredWithError(factReq, &SystemDMissingArgument)
			facts = append(facts, fact)
			continue
		}

		properties, err := g.dbusConnnector.GetUnitPropertiesContext(ctx, factReq.Argument)
		if ctx.Err() != nil {
			break
		}
		if err != nil {
			gatheringError := SystemDUnitError.
				Wrap(fmt.Sprintf("argument %s", factReq.Argument)).
				Wrap(err.Error())
			slog.Error(gatheringError.Error())
			facts = append(facts, entities.NewFactGatheredWithError(factReq, gatheringError))
			continue
		}

		factValue, err := unitPropertiesToFactValue(properties)
		if err != nil {
			gatheringError := SystemDDecodingError.
				Wrap(fmt.Sprintf("argument %s", factReq.Argument)).
				Wrap(err.Error())
			slog.Error(gatheringError.Error())
			facts = append(facts, entities.NewFactGatheredWithError(factReq, gatheringError))
			continue
		}

		facts = append(facts, entities.NewFactGatheredWithRequest(factReq, factValue))

	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	slog.Info("Requested facts gathered", "gatherer", SystemDGathererName, "version", "v2")
	return facts, nil

}

func unitPropertiesToFactValue(properties map[string]interface{}) (entities.FactValue, error) {
	// all the values are always present in the map, no need for checking if they exist
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
