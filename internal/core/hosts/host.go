package hosts

import (
	"context"
	"log/slog"

	"github.com/trento-project/agent/internal/core/hosts/systemd"
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
	dbus, err := systemd.NewDbusConnector(ctx)
	if err != nil {
		slog.Error("Error while creating dbus connection", "error", err)
	}
	return GetSystemdUnitsStatusWithCustomDbus(ctx, dbus, units)
}

func GetSystemdUnitsStatusWithCustomDbus(
	ctx context.Context,
	dbus systemd.DbusConnector,
	units []string,
) []UnitInfo {
	unitsInfo := defaultUnitInfo(units)

	if dbus == nil {
		return unitsInfo
	}

	defer dbus.Close()

	for idx, unit := range unitsInfo {
		unitProperties, err := dbus.GetUnitPropertiesContext(ctx, unit.Name)
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
		unitsInfo[idx].UnitFileState = stringUnitFileState
	}

	return unitsInfo
}
