/*
Based on
- https://cloud.google.com/compute/docs/metadata/overview
- https://cloud.google.com/compute/docs/metadata/querying-metadata
*/

package cloud

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const (
	gcpMetadataUrl          = "http://metadata.google.internal/computeMetadata/v1/"
	gcpMetadataFlavorHeader = "Google"
)

type GcpMetadata struct {
	Instance GcpInstance `json:"instance,omitempty"`
	Project  GcpProject  `json:"project,omitempty"`
}

type GcpInstance struct {
	Disks             []GcpDisk             `json:"disks,omitempty"`
	Image             string                `json:"image,omitempty"`
	MachineType       string                `json:"machineType,omitempty"`
	Name              string                `json:"name,omitempty"`
	NetworkInterfaces []GcpNetworkInterface `json:"networkInterfaces,omitempty"`
	Zone              string                `json:"zone,omitempty"`
}

type GcpDisk struct {
	DeviceName string `json:"deviceName,omitempty"`
	Index      int    `json:"index,omitempty"`
}

type GcpNetworkInterface struct {
	Network string `json:"network,omitempty"`
}

type GcpProject struct {
	ProjectID string `json:"projectId,omitempty"`
}

func NewGcpMetadata() (*GcpMetadata, error) {
	var err error
	m := &GcpMetadata{}

	req, _ := http.NewRequest(http.MethodGet, gcpMetadataUrl, nil)
	req.Header.Add("Metadata-Flavor", gcpMetadataFlavorHeader)

	q := req.URL.Query()
	q.Add("recursive", "true")
	req.URL.RawQuery = q.Encode()

	log.Debug("Requesting GCP metadata...")

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var pjson bytes.Buffer
	err = json.Indent(&pjson, body, "", " ")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Debugln(string(pjson.Bytes()))

	err = json.Unmarshal(body, m)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return m, nil
}
