package subscription

import (
	"encoding/json"
	"log/slog"

	"github.com/pkg/errors"
	"github.com/trento-project/agent/pkg/utils"
)

type Subscriptions []*Subscription

type Subscription struct {
	Identifier string `json:"identifier,omitempty"`
	Version    string `json:"version,omitempty"`
	Arch       string `json:"arch,omitempty"`
	Status     string `json:"status,omitempty"`
	// RegCode string `json:"regcode,omitempty"`
	StartsAt           string `json:"starts_at,omitempty"`
	ExpiresAt          string `json:"expires_at,omitempty"`
	SubscriptionStatus string `json:"subscription_status,omitempty"`
	Type               string `json:"type,omitempty"`
}

func NewSubscriptions(commandExecutor utils.CommandExecutor) (Subscriptions, error) {
	var subs Subscriptions

	slog.Info("Identifying the SUSE subscription details...")
	output, err := commandExecutor.Exec("SUSEConnect", "-s")
	if err != nil {
		return nil, err
	}

	slog.Debug("SUSEConnect output", "output", string(output))

	err = json.Unmarshal(output, &subs)
	if err != nil {
		return nil, errors.Wrap(err, "error while decoding the subscription details")
	}
	slog.Info("Subscription discovered", "entries", len(subs))

	return subs, nil
}
