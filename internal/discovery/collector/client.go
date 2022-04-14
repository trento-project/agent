package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/afero"
)

type Client interface {
	Publish(discoveryType string, payload interface{}) error
	Heartbeat() error
}

type client struct {
	config     *Config
	agentID    string
	httpClient *http.Client
}

type Config struct {
	ServerUrl string
	ApiKey    string
}

const machineIdPath = "/etc/machine-id"

var fileSystem = afero.NewOsFs()

func NewCollectorClient(config *Config) (*client, error) {
	var err error

	httpClient := &http.Client{}

	machineIDBytes, err := afero.ReadFile(fileSystem, machineIdPath)

	if err != nil {
		return nil, err
	}

	machineID := strings.TrimSpace(string(machineIDBytes))

	agentID := uuid.NewSHA1(TrentoNamespace, []byte(machineID))

	return &client{
		config:     config,
		httpClient: httpClient,
		agentID:    agentID.String(),
	}, nil
}

func (c *client) Publish(discoveryType string, payload interface{}) error {
	log.Debugf("Sending %s to data collector", discoveryType)

	requestBody, err := json.Marshal(map[string]interface{}{
		"agent_id":       c.agentID,
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

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf(
			"something wrong happened while publishing data to the collector. Status: %d, Agent: %s, discovery: %s",
			resp.StatusCode, c.agentID, discoveryType)
	}

	return nil
}

func (c *client) Heartbeat() error {
	url := fmt.Sprintf("%s/api/hosts/%s/heartbeat", c.config.ServerUrl, c.agentID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	c.enrichRequest(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("server responded with status code %d while sending heartbeat", resp.StatusCode)
	}

	return nil
}

func (c *client) enrichRequest(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Trento-apiKey", c.config.ApiKey)
}
