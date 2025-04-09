/*
Based on
- https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-retrieval.html
- https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-categories.html
*/

package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	awsMetadataURL                = "http://169.254.169.254/latest/"
	awsMetadataResource           = "meta-data"
	metadataTokenTTLHeader string = "X-aws-ec2-metadata-token-ttl-seconds" //nolint
	metadataTokenHeader    string = "X-aws-ec2-metadata-token"             //nolint

	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/configuring-instance-metadata-service.html
	// (TTL) for the token, in seconds, up to a maximum of six hours (21,600 seconds).
	awsEC2MetadataTokenTTL int = 60 * 2 // 2 minutes should be more than enough for the curren discovery loop
)

type AWSMetadata struct {
	AmiID               string              `json:"ami-id"`
	BlockDeviceMapping  map[string]string   `json:"block-device-mapping"`
	IdentityCredentials IdentityCredentials `json:"identity-credentials"`
	InstanceID          string              `json:"instance-id"`
	InstanceType        string              `json:"instance-type,omitempty"`
	Network             AWSNetwork          `json:"network"`
	Placement           Placement           `json:"placement"`
}

type IdentityCredentials struct {
	EC2 struct {
		Info struct {
			AccountID string `json:"AccountId"`
		} `json:"info"`
	} `json:"ec2"`
}

type AWSNetwork struct {
	Interfaces struct {
		Macs map[string]MacEntry `json:"macs"`
	} `json:"interfaces"`
}

type MacEntry struct {
	VpcID string `json:"vpc-id"`
}

type Placement struct {
	AvailabilityZone string `json:"availability-zone"`
	Region           string `json:"region"`
}

func NewAWSMetadata(ctx context.Context, client HTTPClient) (*AWSMetadata, error) {
	var err error
	awsMetadata := &AWSMetadata{
		AmiID:              "",
		BlockDeviceMapping: map[string]string{},
		IdentityCredentials: IdentityCredentials{
			EC2: struct {
				Info struct {
					AccountID string "json:\"AccountId\""
				} "json:\"info\""
			}{
				Info: struct {
					AccountID string "json:\"AccountId\""
				}{
					AccountID: "",
				},
			},
		},
		InstanceID:   "",
		InstanceType: "",
		Network: AWSNetwork{
			Interfaces: struct {
				Macs map[string]MacEntry "json:\"macs\""
			}{
				Macs: map[string]MacEntry{},
			},
		},
		Placement: Placement{
			AvailabilityZone: "",
			Region:           "",
		},
	}

	token, err := requestMetadataToken(ctx, client)
	if err != nil {
		return nil, err
	}

	firstElementsList := []string{fmt.Sprintf("%s/", awsMetadataResource)}
	metadata, err := buildAWSMetadata(ctx, client, awsMetadataURL, firstElementsList, token)
	if err != nil {
		return nil, err
	}

	jsonMetadata, err := json.Marshal(metadata[awsMetadataResource])
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonMetadata, awsMetadata)
	if err != nil {
		return nil, err
	}

	return awsMetadata, err
}

func requestMetadataToken(ctx context.Context, client HTTPClient) (string, error) {
	log.Debug("Fetching IMDS token...")

	url := fmt.Sprintf("%sapi/token", awsMetadataURL)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, nil)

	req.Header.Add(metadataTokenTTLHeader, strconv.Itoa(awsEC2MetadataTokenTTL))

	resp, err := client.Do(req)
	if err != nil {
		log.Debugf("An error occurred while fetching IMDS token: %s", err)
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("failed to fetch metadata token: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	log.Debug("Metadata token fetched successfully")

	return string(body), nil
}

func buildAWSMetadata(
	ctx context.Context,
	client HTTPClient,
	url string,
	elements []string,
	token string,
) (map[string]any, error) {
	metadata := make(map[string]any)

	for _, element := range elements {
		if strings.TrimSpace(element) == "" {
			continue
		}
		newURL := url + element

		response, err := requestMetadata(ctx, client, newURL, token)
		if err != nil {
			return metadata, err
		}

		if element[len(element)-1:] == "/" {
			currentElement := strings.Trim(element, "/")
			newElements := strings.Split(fmt.Sprintf("%v", response), "\n")

			metadata[currentElement], err = buildAWSMetadata(ctx, client, newURL, newElements, token)
			if err != nil {
				return nil, err
			}
		} else {
			metadata[element] = response
		}
	}

	return metadata, nil
}

func requestMetadata(
	ctx context.Context,
	client HTTPClient,
	url string,
	token string,
) (any, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Add(metadataTokenHeader, token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, errors.Errorf("failed to fetch AWS metadata: %s", resp.Status)
	}

	// The metadata endpoint may return json elements
	if json.Valid(body) {
		var jsonData any
		err := json.Unmarshal(body, &jsonData)
		return jsonData, err
	}
	return string(body), nil
}
