// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type Client interface {
	Publish(ctx context.Context, discoveryType string, payload any) error
	Heartbeat(ctx context.Context) error
}

type Collector struct {
	config     *Config
	httpClient *http.Client
}

type Config struct {
	AgentID   string
	ServerURL string
	APIKey    string
}

func NewCollectorClient(config *Config, httpClient *http.Client) *Collector {
	return &Collector{
		config:     config,
		httpClient: httpClient,
	}
}

func (c *Collector) Publish(ctx context.Context, discoveryType string, payload any) error {
	slog.Debug("Sending to data collector", "discoveryType", discoveryType)

	requestBody, err := json.Marshal(map[string]any{
		"agent_id":       c.config.AgentID,
		"discovery_type": discoveryType,
		"payload":        payload,
	})
	if err != nil {
		return err
	}

	url := c.config.ServerURL + "/api/v1/collect"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	c.enrichRequest(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf(
			"something wrong happened while publishing data to the collector. Status: %d, Agent: %s, discovery: %s",
			resp.StatusCode, c.config.AgentID, discoveryType)
	}

	return nil
}

func (c *Collector) Heartbeat(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/hosts/%s/heartbeat", c.config.ServerURL, c.config.AgentID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	c.enrichRequest(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("server responded with status code %d while sending heartbeat", resp.StatusCode)
	}

	return nil
}

func (c *Collector) enrichRequest(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Trento-Apikey", c.config.APIKey)
}
