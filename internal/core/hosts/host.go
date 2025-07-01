package hosts

import (
	"context"
	"log/slog"

	"github.com/trento-project/agent/internal/core/hosts/systemd"
)

type UnitInfo struct {
	UnitFileState string `json:"unit_file_state"`
}

type SystemdUnitsStatus map[string]UnitInfo

func DefaultUnitInfo(units ...string) SystemdUnitsStatus {
	defaultInfo := make(SystemdUnitsStatus)
	if len(units) == 0 {
		units = []string{"pacemaker.service"}
	}
	for _, unit := range units {
		defaultInfo[unit] = UnitInfo{
			UnitFileState: "unknown",
		}
	}
	return defaultInfo
}

func GetSystemdUnitsStatus(ctx context.Context, units ...string) SystemdUnitsStatus {
	dbus, err := systemd.NewDbusConnector(ctx)
	if err != nil {
		slog.Error("Error while creating dbus connection", "error", err)
	}
	return getSystemdUnitsStatus(ctx, dbus, units...)
}

func GetSystemdUnitsStatusWithCustomDbus(
	ctx context.Context,
	dbus systemd.DbusConnector,
	units ...string,
) SystemdUnitsStatus {
	return getSystemdUnitsStatus(ctx, dbus, units...)
}

func getSystemdUnitsStatus(ctx context.Context, dbus systemd.DbusConnector, units ...string) SystemdUnitsStatus {
	unitsInfo := DefaultUnitInfo(units...)

	if dbus == nil {
		return unitsInfo
	}

	defer dbus.Close()

	for unit := range unitsInfo {
		unitProperties, err := dbus.GetUnitPropertiesContext(ctx, unit)
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
		unitsInfo[unit] = UnitInfo{UnitFileState: stringUnitFileState}
	}

	return unitsInfo
}
