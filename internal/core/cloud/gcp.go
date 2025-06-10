/*
Based on
- https://cloud.google.com/compute/docs/metadata/overview
- https://cloud.google.com/compute/docs/metadata/querying-metadata
*/

package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"log/slog"
)

const (
	gcpMetadataURL          = "http://metadata.google.internal/computeMetadata/v1/"
	gcpMetadataFlavorHeader = "Google"
)

type GCPMetadata struct {
	Instance GCPInstance `json:"instance,omitempty"`
	Project  GCPProject  `json:"project,omitempty"`
}

type GCPInstance struct {
	Disks             []GCPDisk             `json:"disks,omitempty"`
	Image             string                `json:"image,omitempty"`
	MachineType       string                `json:"machineType,omitempty"`
	Name              string                `json:"name,omitempty"`
	NetworkInterfaces []GCPNetworkInterface `json:"networkInterfaces,omitempty"`
	Zone              string                `json:"zone,omitempty"`
}

type GCPDisk struct {
	DeviceName string `json:"deviceName,omitempty"`
	Index      int    `json:"index,omitempty"`
}

type GCPNetworkInterface struct {
	Network string `json:"network,omitempty"`
}

type GCPProject struct {
	ProjectID string `json:"projectId,omitempty"`
}

func NewGCPMetadata(ctx context.Context, client HTTPClient) (*GCPMetadata, error) {
	var err error
	m := &GCPMetadata{
		Instance: GCPInstance{
			Disks:             []GCPDisk{},
			Image:             "",
			MachineType:       "",
			Name:              "",
			NetworkInterfaces: []GCPNetworkInterface{},
			Zone:              "",
		},
		Project: GCPProject{
			ProjectID: "",
		},
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, gcpMetadataURL, nil)
	req.Header.Add("Metadata-Flavor", gcpMetadataFlavorHeader)

	q := req.URL.Query()
	q.Add("recursive", "true")
	req.URL.RawQuery = q.Encode()

	slog.Debug("Requesting GCP metadata...")

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("failed to get GCP metadata", "error", err.Error())
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to read GCP metadata", "error", err.Error())
		return nil, err
	}

	var pjson bytes.Buffer
	err = json.Indent(&pjson, body, "", " ")
	if err != nil {
		slog.Error("failed to indent GCP metadata", "error", err.Error())
		return nil, err
	}
	slog.Debug(pjson.String())

	err = json.Unmarshal(body, m)
	if err != nil {
		slog.Error("failed to unmarshal GCP metadata", "error", err.Error())
		return nil, err
	}

	return m, nil
}
