package systemd

import (
	"context"

	"github.com/coreos/go-systemd/v22/dbus"
)

// DbusConnector acts as an abstract interface for the dbus functionalities
// exposed by the package "github.com/coreos/go-systemd/v22/dbus"
type DbusConnector interface {
	GetUnitPropertiesContext(ctx context.Context, unit string) (map[string]any, error)
	ListUnitsByNamesContext(ctx context.Context, units []string) ([]dbus.UnitStatus, error)
	// NewWithContext establishes a connection to any available bus and authenticates.
	// Callers should call Close() when done with the connection.
	// see https://pkg.go.dev/github.com/coreos/go-systemd/v22@v22.5.0/dbus#NewWithContext
	Close()
}

func NewDbusConnector(ctx context.Context) (DbusConnector, error) {
	// the created connection does implement the DbusConnector interface, hence it can be returned as such
	dbusConnection, err := dbus.NewWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return dbusConnection, nil
}
