package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Client interface {
	Publish(discoveryType string, payload interface{}) error
	Heartbeat() error
}

type client struct {
	config     *Config
	httpClient *http.Client
}

type Config struct {
	AgentID   string
	ServerUrl string
	ApiKey    string
}

func NewCollectorClient(config *Config) *client {
	return &client{
		config:     config,
		httpClient: &http.Client{},
	}
}

func (c *client) Publish(discoveryType string, payload interface{}) error {
	log.Debugf("Sending %s to data collector", discoveryType)

	requestBody, err := json.Marshal(map[string]interface{}{
		"agent_id":       c.config.AgentID,
		"discovery_type": discoveryType,
		"payload":        payload,
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/collect", c.config.ServerUrl)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
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

func (c *client) Heartbeat() error {
	url := fmt.Sprintf("%s/api/hosts/%s/heartbeat", c.config.ServerUrl, c.config.AgentID)

	req, err := http.NewRequest("POST", url, nil)
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

func (c *client) enrichRequest(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Trento-apiKey", c.config.ApiKey)
}
