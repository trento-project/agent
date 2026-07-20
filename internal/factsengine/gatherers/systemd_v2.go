// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package gatherers

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/trento-project/agent/internal/core/dbus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

// nolint:gochecknoglobals
//nolint:gochecknoglobals
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
	dbusConnector dbus.Connector
	initialized   bool
}

type systemdUnitStatus struct {
	ActiveState      any `json:"active_state"`
	Description      any `json:"description"`
	ID               any `json:"id"`
	LoadState        any `json:"load_state"`
	NeedDaemonReload any `json:"need_daemon_reload"`
	UnitFilePreset   any `json:"unit_file_preset"`
	UnitFileState    any `json:"unit_file_state"`
}

func NewDefaultSystemDGathererV2() *SystemDGathererV2 {
	ctx := context.Background()

	dbusConnector, err := dbus.NewConnector(ctx)
	if err != nil {
		slog.Error("Error initializing dbus", "error", err)

		return &SystemDGathererV2{
			dbusConnector: nil,
			initialized:   false,
		}
	}

	return NewSystemDGathererV2(dbusConnector, true)
}

func NewSystemDGathererV2(conn dbus.Connector, initialized bool) *SystemDGathererV2 {
	return &SystemDGathererV2{
		dbusConnector: conn,
		initialized:   initialized,
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

		properties, err := g.dbusConnector.GetUnitPropertiesContext(ctx, factReq.Argument)

		if ctx.Err() != nil {
			break
		}

		if err != nil {
			gatheringError := SystemDUnitError.
				Wrap("argument " + factReq.Argument).
				Wrap(err.Error())
			slog.Error(gatheringError.Error())
			facts = append(facts, entities.NewFactGatheredWithError(factReq, gatheringError))

			continue
		}

		factValue, err := unitPropertiesToFactValue(properties)
		if err != nil {
			gatheringError := SystemDDecodingError.
				Wrap("argument " + factReq.Argument).
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

func unitPropertiesToFactValue(properties map[string]any) (entities.FactValue, error) {
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

	var unmarshalled map[string]any

	err = json.Unmarshal(marshalled, &unmarshalled)
	if err != nil {
		return nil, err
	}

	return entities.NewFactValue(unmarshalled, entities.WithStringConversion())
}
