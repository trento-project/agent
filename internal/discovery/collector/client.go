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

type Collector struct {
	config     *Config
	httpClient *http.Client
}

type Config struct {
	AgentID   string
	ServerURL string
	APIKey    string
}

func NewCollectorClient(config *Config) *Collector {
	return &Collector{
		config:     config,
		httpClient: http.DefaultClient,
	}
}

func (c *Collector) Publish(discoveryType string, payload interface{}) error {
	log.Debugf("Sending %s to data collector", discoveryType)

	requestBody, err := json.Marshal(map[string]interface{}{
		"agent_id":       c.config.AgentID,
		"discovery_type": discoveryType,
		"payload":        payload,
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/collect", c.config.ServerURL)

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

func (c *Collector) Heartbeat() error {
	url := fmt.Sprintf("%s/api/v1/hosts/%s/heartbeat", c.config.ServerURL, c.config.AgentID)

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

func (c *Collector) enrichRequest(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Trento-apiKey", c.config.APIKey)
}
