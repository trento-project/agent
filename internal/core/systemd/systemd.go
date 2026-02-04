package systemd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/trento-project/agent/internal/core/dbus"
)

type UnitInfo struct {
	Name          string `json:"name"`
	UnitFileState string `json:"unit_file_state"`
}

type Systemd interface {
	Enable(ctx context.Context, service string) error
	Disable(ctx context.Context, service string) error
	IsActive(ctx context.Context, service string) (bool, error)
	IsEnabled(ctx context.Context, service string) (bool, error)
	GetUnitsInfo(ctx context.Context, units []string) []UnitInfo
	Close()
}

type Connector struct {
	dbusConnection dbus.Connector
	logger         *slog.Logger
}

type ConnectorOption func(*Connector)

type Loader interface {
	NewSystemd(ctx context.Context, options ...ConnectorOption) (Systemd, error)
}

type defaultSystemdLoader struct{}

func (d *defaultSystemdLoader) NewSystemd(
	ctx context.Context,
	options ...ConnectorOption,
) (Systemd, error) {
	return NewSystemd(ctx, options...)
}

func NewDefaultSystemdLoader() Loader {
	return &defaultSystemdLoader{}
}

func WithCustomDbusConnector(dbusConnection dbus.Connector) ConnectorOption {
	return func(s *Connector) {
		s.dbusConnection = dbusConnection
	}
}

func WithCustomLogger(logger *slog.Logger) ConnectorOption {
	return func(s *Connector) {
		s.logger = logger
	}
}

func NewSystemd(ctx context.Context, options ...ConnectorOption) (Systemd, error) {
	systemdInstance := &Connector{
		logger: slog.Default(),
	}

	for _, opt := range options {
		opt(systemdInstance)
	}

	if systemdInstance.dbusConnection != nil {
		return systemdInstance, nil
	}

	dbusConnection, err := dbus.NewConnector(ctx)
	if err != nil {
		systemdInstance.logger.Error("failed to create dbus connection", "error", err)
		return nil, err
	}
	systemdInstance.dbusConnection = dbusConnection

	return systemdInstance, nil
}

func (s *Connector) Enable(ctx context.Context, service string) error {
	_, _, err := s.dbusConnection.EnableUnitFilesContext(ctx, []string{service}, false, true)
	if err != nil {
		s.logger.Error("failed to enable service", "service", service, "error", err)
		return fmt.Errorf("failed to enable service %s: %w", service, err)
	}

	return s.reload(ctx, service)
}

func (s *Connector) Disable(ctx context.Context, service string) error {
	_, err := s.dbusConnection.DisableUnitFilesContext(ctx, []string{service}, false)
	if err != nil {
		s.logger.Error("failed to disable service", "service", service, "error", err)
		return fmt.Errorf("failed to disable service %s: %w", service, err)
	}

	return s.reload(ctx, service)
}

func (s *Connector) IsActive(ctx context.Context, service string) (bool, error) {
	activeState, err := s.getUnitProperty(ctx, service, "ActiveState")
	if err != nil {
		return false, err
	}

	return activeState == "active", nil
}

func (s *Connector) IsEnabled(ctx context.Context, service string) (bool, error) {
	unitFileState, err := s.getUnitProperty(ctx, service, "UnitFileState")
	if err != nil {
		return false, err
	}

	return unitFileState == "enabled", nil
}

func (s *Connector) GetUnitsInfo(
	ctx context.Context,
	units []string,
) []UnitInfo {
	unitsInfo := []UnitInfo{}
	for _, unit := range units {
		unitsInfo = append(unitsInfo, UnitInfo{
			Name:          unit,
			UnitFileState: "unknown",
		})
	}

	for idx, unit := range unitsInfo {
		unitProperties, err := s.dbusConnection.GetUnitPropertiesContext(ctx, unit.Name)
		if err != nil {
			s.logger.Error("Error getting systemd unit properties", "unit", unit, "error", err)
			continue
		}
		unitFileState, ok := unitProperties["UnitFileState"]
		if !ok {
			s.logger.Warn("UnitFileState not found in properties", "unit", unit, "properties", unitProperties)
			continue
		}
		stringUnitFileState, ok := unitFileState.(string)
		if !ok {
			s.logger.Warn("UnitFileState is not a string", "unit", unit, "value", unitFileState)
			continue
		}
		if stringUnitFileState == "" {
			s.logger.Warn("UnitFileState is empty, service probably not installed", "unit", unit, "value", stringUnitFileState)
			continue
		}
		unitsInfo[idx].UnitFileState = stringUnitFileState
	}

	return unitsInfo
}

func (s *Connector) Close() {
	s.dbusConnection.Close()
}

func (s *Connector) reload(ctx context.Context, service string) error {
	err := s.dbusConnection.ReloadContext(ctx)
	if err != nil {
		s.logger.Error("failed to reload service", "service", service, "error", err)
		return fmt.Errorf("failed to reload service %s: %w", service, err)
	}
	return nil
}

func (s *Connector) getUnitProperty(ctx context.Context, unit string, propertyName string) (string, error) {
	property, err := s.dbusConnection.GetUnitPropertyContext(ctx, unit, propertyName)
	if err != nil {
		s.logger.Error("failed to get property for service", "service", unit, "error", err)
		return "", fmt.Errorf("failed to get property %s for service %s: %w", propertyName, unit, err)
	}

	value, ok := property.Value.Value().(string)
	if !ok {
		s.logger.Error("unexpected type for service", "service", unit,
			"type", fmt.Sprintf("%T", property.Value.Value()))
		return "", fmt.Errorf("unexpected type for service %s: %T",
			unit, property.Value.Value())
	}

	return value, nil
}
