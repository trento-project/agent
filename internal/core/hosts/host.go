package hosts

import (
	"context"
	"log/slog"

	"github.com/trento-project/agent/internal/core/dbus"
)

type UnitInfo struct {
	Name          string `json:"name"`
	UnitFileState string `json:"unit_file_state"`
}

func defaultUnitInfo(units []string) []UnitInfo {
	defaultInfo := []UnitInfo{}
	for _, unit := range units {
		defaultInfo = append(defaultInfo, UnitInfo{
			Name:          unit,
			UnitFileState: "unknown",
		})
	}
	return defaultInfo
}

func GetSystemdUnitsStatus(ctx context.Context, units []string) []UnitInfo {
	dbusConnector, err := dbus.NewConnector(ctx)
	if err != nil {
		slog.Error("Error while creating dbus connection", "error", err)
	}
	return GetSystemdUnitsStatusWithCustomDbus(ctx, dbusConnector, units)
}

func GetSystemdUnitsStatusWithCustomDbus(
	ctx context.Context,
	dbusConnector dbus.Connector,
	units []string,
) []UnitInfo {
	unitsInfo := defaultUnitInfo(units)

	if dbusConnector == nil {
		return unitsInfo
	}

	defer dbusConnector.Close()

	for idx, unit := range unitsInfo {
		unitProperties, err := dbusConnector.GetUnitPropertiesContext(ctx, unit.Name)
		if err != nil {
			slog.Error("Error getting systemd unit properties", "unit", unit, "error", err)
			continue
		}
		unitFileState, ok := unitProperties["UnitFileState"]
		if !ok {
			slog.Warn("UnitFileState not found in properties", "unit", unit, "properties", unitProperties)
			continue
		}
		stringUnitFileState, ok := unitFileState.(string)
		if !ok {
			slog.Warn("UnitFileState is not a string", "unit", unit, "value", unitFileState)
			continue
		}
		if stringUnitFileState == "" {
			slog.Warn("UnitFileState is empty, service probably not installed", "unit", unit, "value", stringUnitFileState)
			continue
		}
		unitsInfo[idx].UnitFileState = stringUnitFileState
	}

	return unitsInfo
}
